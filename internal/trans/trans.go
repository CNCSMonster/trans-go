package trans

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cncsmonster/trans-go/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type EnTextKind int

func EnTextKindFromStr(s string) (EnTextKind, error) {
	switch s {
	case "INVALID":
		return INVALID, nil
	case "WORD":
		return WORD, nil
	case "PHRASE":
		return PHRASE, nil
	case "SENTENCE":
		return SENTENCE, nil
	case "Paragraph":
		return Paragraph, nil
	default:
		return INVALID, fmt.Errorf("invalid text kind: %s", s)
	}
}

const (
	INVALID EnTextKind = iota
	WORD
	PHRASE
	SENTENCE
	Paragraph
)

const (
	systemPromptAuto = `
	你是一个翻译专家，你的任务是将我给出的英文翻译成中文,要求做到信雅达;
	如果我给出的英文是单词:
	你首先要给出该单词最常用的解释以及该单词的音标,然后给出一个英文的解释,同时你需要使用该解释给出英文例句以及其翻译;
	格式如下:
	{中文翻译}  {音标}
	{英文解释}
	例句:
	1. {英文例句} {例句的中文翻译}
	...;
	如果该单词有其他解释和词性,则继续补充说明;
	`

	systemPromptWord = `
	你是一个翻译专家，你的任务是将我给出的英文单词翻译成中文,要求做到信雅达;
	你首先要给出该单词最常用的解释以及该单词的音标, 然后给出一个英文的解释, 同时你需要使用该解释给出英文例句以及其翻译;
	格式如下:
	{中文翻译}  {音标}
	{英文解释}
	例句:
	1. {英文例句} {例句的中文翻译}
	...;
	如果该单词有其他解释和词性,则继续适当补充说明;
	`
	systemPromptPhrase = `
	你是一个翻译专家，你的任务是将我给出的英文短语翻译成中文,要求做到信雅达;
	你需要给出该短语的中文翻译，以及用法说明和例句;
	格式如下:
	{中文翻译}
	用法说明：
	{用法说明}
	例句:
	1. {英文例句} {例句的中文翻译}
	...
	`
	systemPromptSentence = `
	你是一个翻译专家，你的任务是将我给出的英文句子翻译成中文,要求做到信雅达;
	你需要给出该句子的中文翻译，并解释句子中的重要语法点或词组用法;
	格式如下:
	译文：{中文翻译}
	要点解：
	1. {重要语法点或词组用法解释}
	...
	`
	systemPromptParagraph = `
	你是一个翻译专家，你的任务是将我给出的英文段落翻译成中文,要求做到信雅达;
	你需要给出该段落的中文翻译，并总结段落的主要内容和写作特点;
	格式如下:
	译文：
	{中文翻译}
	
	内容要点：
	1. {主要内容概述}
	2. {写作特点分析}
	...
	`
)

type Translator struct {
	Client *openai.Client
	Config *config.Config
}

func NewTranslator(conf config.Config) (Translator, error) {
	if conf.ApiKey == "" {
		return Translator{}, fmt.Errorf("API key is required")
	}
	if conf.BaseUrl == "" {
		conf.BaseUrl = "https://api.openai.com/v1" // 设默认值
	}

	return Translator{
		Client: openai.NewClient(option.WithBaseURL(conf.BaseUrl), option.WithAPIKey(conf.ApiKey)),
		Config: &conf,
	}, nil
}

func (translator Translator) Translate(texts ...string) error {
	if translator.Config.Optimize > 0 {
		return translator.TranslateO1(texts...)
	} else {
		return translator.TranslateAuto(texts...)
	}
}

func (translator Translator) TranslateO1(texts ...string) error {
	for _, text := range texts {
		if err := translator.translateSingle(text); err != nil {
			return err
		}
	}
	return nil
}

func (translator Translator) AnalyzeKind(text string) (EnTextKind, error) {
	systemPrompt := `
	你是一个英文文本类型检测器，你将精确的根据自己的英文知识分析出发送给你的英文文本的类型。
	你只能回复以下几种类型之一：
	WORD,PHRASE,SENTENCE,Paragraph
	这些类型分别意味着:
	WORD: 语言或写作中的一个独立且有意义的元素，通常与其他词一起构成句子。
	PHRASE: 一组作为概念单位的小词群，通常构成从句的一个组成部分。
	SENTENCE: 一组完整的词语，通常包含主语和谓语，表达陈述、疑问、感叹或命令。
	Paragraph: 文章中的一个独立部分，通常围绕一个主题展开，并通过新行、缩进或编号来标识。
	`
	ctx := context.Background()
	conf := translator.Config
	resp, err := translator.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(text),
		}),
		Seed:        openai.Int(1),
		Model:       openai.F(conf.Model),
		Temperature: openai.Float(0),
	})
	if err != nil {
		return INVALID, err
	}

	ans := resp.Choices[0].Message.Content
	ans = strings.TrimSpace(ans)

	if translator.Config.Verbose {
		fmt.Printf("=== 文本分析 ===\n")
		fmt.Printf("输入: %q\n", text)
		fmt.Printf("类型: %s\n", ans)
		fmt.Printf("==============\n")
	}

	return EnTextKindFromStr(ans)
}

func (translator Translator) TranslateAuto(texts ...string) error {
	ctx := context.Background()
	systemPrompt := systemPromptAuto
	for _, text := range texts {
		if err := translator.streamingTranslate(ctx, systemPrompt, text); err != nil {
			return fmt.Errorf("translate auto failed: %w", err)
		}
	}
	return nil
}

func (translator Translator) TranslateWord(words ...string) error {
	ctx := context.Background()
	systemPrompt := systemPromptWord
	for _, word := range words {
		if err := translator.streamingTranslate(ctx, systemPrompt, word); err != nil {
			return fmt.Errorf("translate word failed: %w", err)
		}
	}
	return nil
}

func (translator Translator) streamingTranslate(ctx context.Context, systemPrompt string, text string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client := translator.Client
	conf := translator.Config

	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(text),
		}),
		Seed:        openai.Int(1),
		Model:       openai.F(conf.Model),
		Temperature: openai.Float(conf.Temperature),
	})

	for stream.Next() {
		if err := stream.Err(); err != nil {
			return fmt.Errorf("streaming error: %w", err)
		}
		current := stream.Current()
		if len(current.Choices) == 0 {
			break
		}
		for _, c := range current.Choices {
			chunk := c.Delta.Content
			fmt.Printf("%s", chunk)
		}
	}
	fmt.Println()
	return nil
}

func (translator Translator) TranslatePhrase(phrases ...string) error {
	ctx := context.Background()
	systemPrompt := systemPromptPhrase
	for _, phrase := range phrases {
		if err := translator.streamingTranslate(ctx, systemPrompt, phrase); err != nil {
			return fmt.Errorf("translate phrase failed: %w", err)
		}
	}
	return nil
}

func (translator Translator) TranslateSentence(sentences ...string) error {
	ctx := context.Background()
	systemPrompt := systemPromptSentence
	for _, sentence := range sentences {
		if err := translator.streamingTranslate(ctx, systemPrompt, sentence); err != nil {
			return fmt.Errorf("translate sentence failed: %w", err)
		}
	}
	return nil
}

func (translator Translator) TranslateParagraph(paragraphs ...string) error {
	ctx := context.Background()
	systemPrompt := systemPromptParagraph
	for _, paragraph := range paragraphs {
		if err := translator.streamingTranslate(ctx, systemPrompt, paragraph); err != nil {
			return fmt.Errorf("translate paragraph failed: %w", err)
		}
	}
	return nil
}

func (translator Translator) translateSingle(text string) error {
	kind, err := translator.AnalyzeKind(text)
	if err != nil {
		return fmt.Errorf("analyze text kind failed: %w", err)
	}

	switch kind {
	case WORD:
		return translator.TranslateWord(text)
	case PHRASE:
		return translator.TranslatePhrase(text)
	case SENTENCE:
		return translator.TranslateSentence(text)
	case Paragraph:
		return translator.TranslateParagraph(text)
	default:
		return fmt.Errorf("unsupported text kind: %v", kind)
	}
}
