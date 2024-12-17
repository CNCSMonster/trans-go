package trans

import (
	"context"
	"fmt"

	"github.com/cncsmonster/trans-go/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func Translate(conf *config.Config, toTranslate ...string) {

	client := openai.NewClient(option.WithBaseURL(conf.BaseUrl), option.WithAPIKey(conf.ApiKey))

	ctx := context.Background()

	systemPrompt := `
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
	for _, question := range toTranslate {
		stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(question),
			}),
			Seed:        openai.Int(1),
			Model:       openai.F(conf.Model),
			Temperature: openai.Float(conf.Temperature),
		})
		for stream.Next() {
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
	}
}
