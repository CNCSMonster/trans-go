/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	"github.com/cncsmonster/trans-go/internal/config"
	"github.com/cncsmonster/trans-go/internal/trans"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trans",
	Short: "A simple tool using llm to translate english in terminal",
	Long:  `use - will get input from stdin, input 'EOF' (usually <C+D>) to finish input`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		optimize, _ := cmd.Flags().GetInt("optimize")
		conf := config.NewConfig()
		conf.Optimize = optimize
		conf.Verbose = verbose
		if verbose {
			fmt.Println(conf)
		}

		// 创建 translator 实例
		translator, err := trans.NewTranslator(conf)
		if err != nil {
			log.Fatal(err)
		}

		if slices.Contains(args, "-") || len(args) == 0 {
			stdin, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			if err := translator.Translate(string(stdin)); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := translator.Translate(args...); err != nil {
				log.Fatal(err)
			}
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
	rootCmd.Flags().BoolP("verbose", "v", false, "show detail information")
	rootCmd.Flags().IntP("optimize", "o", 0, "optimization level")
}
