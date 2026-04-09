package cmd

import (
	"fmt"
	"os"

	"github.com/jholm117/hackerrank-cli/internal/api"
	"github.com/jholm117/hackerrank-cli/internal/output"
	"github.com/spf13/cobra"
)

var testsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Manage tests",
}

var testsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		c, err := newClient()
		if err != nil {
			return err
		}

		tests, err := api.PaginateN[api.Test](c, "/tests", nil, limit)
		if err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, tests)
		}

		w := output.NewTableWriter(os.Stdout)
		w.SetHeader([]string{"ID", "NAME", "STATE", "DRAFT", "QUESTIONS"})
		for _, t := range tests {
			w.Append([]string{
				t.ID,
				t.Name,
				t.State,
				fmt.Sprintf("%v", t.Draft),
				fmt.Sprintf("%d", len(t.Questions)),
			})
		}
		w.Render()
		return nil
	},
}

var testsGetCmd = &cobra.Command{
	Use:   "get <test-id>",
	Short: "Show test details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var test api.Test
		if err := c.Get("/tests/"+args[0], nil, &test); err != nil {
			return err
		}

		if flagOutput == "table" {
			w := output.NewTableWriter(os.Stdout)
			w.SetHeader([]string{"FIELD", "VALUE"})
			w.Append([]string{"ID", test.ID})
			w.Append([]string{"Name", test.Name})
			w.Append([]string{"State", test.State})
			w.Append([]string{"Draft", fmt.Sprintf("%v", test.Draft)})
			w.Append([]string{"Duration", fmt.Sprintf("%d min", test.Duration)})
			w.Append([]string{"Questions", fmt.Sprintf("%d", len(test.Questions))})
			w.Append([]string{"Created", test.CreatedAt})
			w.Render()
			return nil
		}

		return output.WriteJSON(os.Stdout, test)
	},
}

func init() {
	testsListCmd.Flags().Int("limit", 20, "Max results to return (0 for all)")
	testsCmd.AddCommand(testsListCmd)
	testsCmd.AddCommand(testsGetCmd)
	rootCmd.AddCommand(testsCmd)
}
