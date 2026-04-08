package cmd

import (
	"github.com/spf13/cobra"
)

var (
	flagToken   string
	flagOutput  string
	flagNoColor bool
)

var rootCmd = &cobra.Command{
	Use:           "hr",
	Short:         "CLI for HackerRank for Work API",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (overrides config and env)")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table or json")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable color output")
}

func Execute() error {
	return rootCmd.Execute()
}
