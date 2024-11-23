/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	"github.com/cncsmonster/tl-go/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"
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
		//var total string
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tl",
	Short: "A simple tool using llm to translate english in terminal",
	Long: `use - will get input from stdin,input 'EOF' (usually <C+D>) to finish input`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		conf := config.NewConfig()
		if verbose {
			fmt.Println(conf)
		}
		if slices.Contains(args, "-") {
			stdin, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			Translate(&conf, string(stdin))
		} else {
			Translate(&conf, args...)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ts.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("verbose", "v", false, "show in for")
}
