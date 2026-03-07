package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/git"
	"github.com/spf13/cobra"
)

var gitContextCmd = &cobra.Command{
	Use:   "git-context [action]",
	Short: "Consolidated git state gathering",
	Long:  "Gather consolidated git context (branch, status, commits, diff, sensitive files) as JSON.\nOptional actions: changed-files, sensitive-check.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGitContext,
}

func init() {
	rootCmd.AddCommand(gitContextCmd)
}

func runGitContext(cmd *cobra.Command, args []string) error {
	action := ""
	if len(args) == 1 {
		action = args[0]
	}

	format := formatFlag
	if format == "" {
		format = "json"
	}

	switch action {
	case "":
		return runGitContextFull(format)
	case "changed-files":
		return runGitContextChangedFiles(format)
	case "sensitive-check":
		return runGitContextSensitiveCheck(format)
	default:
		return fmt.Errorf("unknown action: %s (expected changed-files or sensitive-check)", action)
	}
}

func runGitContextFull(format string) error {
	ctx, err := git.GatherContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(ctx, "", "  ")
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s (expected json)", format)
	}
	return nil
}

func runGitContextChangedFiles(format string) error {
	files, err := git.ChangedFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(files, "", "  ")
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s (expected json)", format)
	}
	return nil
}

func runGitContextSensitiveCheck(format string) error {
	matches, err := git.SensitiveCheck()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(matches, "", "  ")
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s (expected json)", format)
	}
	return nil
}
