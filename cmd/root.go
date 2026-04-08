package cmd

import (
	"fmt"

	"github.com/jholm117/hackerrank-cli/internal/api"
	"github.com/jholm117/hackerrank-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	flagToken   string
	flagOutput  string
	flagNoColor bool
	flagBaseURL string
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
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "API base URL override")
	rootCmd.PersistentFlags().MarkHidden("base-url") //nolint:errcheck
}

func Execute() error {
	return rootCmd.Execute()
}

// newClient creates an API client from resolved token.
func newClient() (*api.Client, error) {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, err
	}
	token := config.ResolveToken(flagToken, cfg)
	if token == "" {
		return nil, fmt.Errorf("not authenticated — run: hr auth login")
	}
	var opts []api.Option
	if flagBaseURL != "" {
		opts = append(opts, api.WithBaseURL(flagBaseURL))
	}
	return api.NewClient(token, opts...), nil
}
