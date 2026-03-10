package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/soudai/saga/internal/config"
	"github.com/soudai/saga/internal/control"
	sagagithub "github.com/soudai/saga/internal/github"
	"github.com/soudai/saga/internal/instructionissue"
)

type issueEnqueueFunc func(ctx context.Context, configPath, repository string, issueNumber int64) (control.TaskResponse, error)

type issueCommandDeps struct {
	writer  sagagithub.IssueWriter
	enqueue issueEnqueueFunc
}

func defaultIssueCommandDeps() issueCommandDeps {
	return issueCommandDeps{
		writer: sagagithub.NewGHCLIIssueWriter(),
		enqueue: func(ctx context.Context, configPath, repository string, issueNumber int64) (control.TaskResponse, error) {
			cfg, err := config.Load(configPath)
			if err != nil {
				return control.TaskResponse{}, err
			}
			return control.NewClient(cfg.Server.SocketPath).Enqueue(ctx, repository, issueNumber)
		},
	}
}

func newIssueCommand(stdin io.Reader, stdout io.Writer, configPath *string) *cobra.Command {
	return newIssueCommandWithDeps(stdin, stdout, configPath, defaultIssueCommandDeps())
}

func newIssueCommandWithDeps(stdin io.Reader, stdout io.Writer, configPath *string, deps issueCommandDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Render or create instruction issues",
	}
	cmd.AddCommand(newIssueDraftCommand(stdin, stdout))
	cmd.AddCommand(newIssueCreateCommand(stdin, stdout, configPath, deps))
	return cmd
}

func newIssueDraftCommand(stdin io.Reader, stdout io.Writer) *cobra.Command {
	var fromFile string
	var title string

	cmd := &cobra.Command{
		Use:   "draft <repository>",
		Short: "Render an instruction issue body locally",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repository, err := normalizeRepository(args[0])
			if err != nil {
				return err
			}

			content, err := readIssueSource(stdin, fromFile)
			if err != nil {
				return err
			}

			rendered, err := instructionissue.Render(content, instructionissue.RenderOptions{
				Repository: repository,
				Title:      title,
			})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(stdout, "%s\n", rendered.Body)
			return err
		},
	}
	cmd.Flags().StringVar(&fromFile, "from-file", "", "path to the source markdown file ('-' to read from stdin)")
	cmd.Flags().StringVar(&title, "title", "", "explicit issue title override")
	return cmd
}

func newIssueCreateCommand(stdin io.Reader, stdout io.Writer, configPath *string, deps issueCommandDeps) *cobra.Command {
	var fromFile string
	var title string
	var labels []string
	var enqueue bool

	cmd := &cobra.Command{
		Use:   "create <repository>",
		Short: "Create an instruction issue on GitHub",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repository, err := normalizeRepository(args[0])
			if err != nil {
				return err
			}

			content, err := readIssueSource(stdin, fromFile)
			if err != nil {
				return err
			}

			rendered, err := instructionissue.Render(content, instructionissue.RenderOptions{
				Repository: repository,
				Title:      title,
			})
			if err != nil {
				return err
			}

			created, err := deps.writer.Create(cmd.Context(), repository, sagagithub.CreateIssueRequest{
				Title:  rendered.Title,
				Body:   rendered.Body,
				Labels: labels,
			})
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintf(stdout, "issue #%d %s\n", created.Number, created.URL); err != nil {
				return err
			}

			if !enqueue {
				return nil
			}

			task, err := deps.enqueue(cmd.Context(), *configPath, repository, created.Number)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(stdout, "task id=%d repo=%s issue=%d state=%s\n", task.ID, task.Repository, task.IssueNumber, task.State)
			return err
		},
	}
	cmd.Flags().StringVar(&fromFile, "from-file", "", "path to the source markdown file ('-' to read from stdin)")
	cmd.Flags().StringVar(&title, "title", "", "explicit issue title override")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "issue label to apply (repeatable)")
	cmd.Flags().BoolVar(&enqueue, "enqueue", false, "register the created issue as a queued local task")
	return cmd
}

func normalizeRepository(value string) (string, error) {
	repository := strings.TrimSpace(value)
	parts := strings.Split(repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("repository must be in owner/repo format")
	}
	return repository, nil
}

func readIssueSource(stdin io.Reader, fromFile string) (string, error) {
	if fromFile == "" {
		return "", fmt.Errorf("--from-file is required")
	}

	var data []byte
	var err error
	if fromFile == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(fromFile)
	}
	if err != nil {
		return "", err
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", fmt.Errorf("instruction issue content is required")
	}
	return content, nil
}
