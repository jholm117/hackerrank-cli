// cmd/interviews.go
package cmd

import (
	"fmt"
	"os"

	"github.com/jholm117/hackerrank-cli/internal/api"
	"github.com/jholm117/hackerrank-cli/internal/output"
	"github.com/spf13/cobra"
)

var interviewsCmd = &cobra.Command{
	Use:   "interviews",
	Short: "Manage interviews",
}

var interviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all interviews",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		interviews, err := api.Paginate[api.Interview](c, "/interviews", nil)
		if err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, interviews)
		}

		w := output.NewTableWriter(os.Stdout)
		w.SetHeader([]string{"ID", "TITLE", "STATUS", "CREATED"})
		for _, iv := range interviews {
			w.Append([]string{iv.ID, iv.Title, iv.Status, iv.CreatedAt})
		}
		w.Render()
		return nil
	},
}

var interviewsGetCmd = &cobra.Command{
	Use:   "get <interview-id>",
	Short: "Show interview details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var iv api.Interview
		if err := c.Get("/interviews/"+args[0], nil, &iv); err != nil {
			return err
		}

		if flagOutput == "table" {
			w := output.NewTableWriter(os.Stdout)
			w.SetHeader([]string{"FIELD", "VALUE"})
			w.Append([]string{"ID", iv.ID})
			w.Append([]string{"Title", iv.Title})
			w.Append([]string{"Status", iv.Status})
			w.Append([]string{"Created", iv.CreatedAt})
			w.Append([]string{"URL", iv.URL})
			w.Render()
			return nil
		}

		return output.WriteJSON(os.Stdout, iv)
	},
}

var interviewsTranscriptCmd = &cobra.Command{
	Use:   "transcript <interview-id>",
	Short: "Get interview transcript",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var transcript api.Transcript
		if err := c.Get("/interviews/"+args[0]+"/transcript", nil, &transcript); err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, transcript)
		}

		for _, msg := range transcript.Messages {
			role := msg.Author
			if msg.Candidate {
				role = fmt.Sprintf("%s (candidate)", msg.Author)
			}
			fmt.Printf("[%s] %s:\n%s\n\n", msg.Timestamp, role, msg.Text)
		}
		return nil
	},
}

func init() {
	interviewsCmd.AddCommand(interviewsListCmd)
	interviewsCmd.AddCommand(interviewsGetCmd)
	interviewsCmd.AddCommand(interviewsTranscriptCmd)
	rootCmd.AddCommand(interviewsCmd)
}
