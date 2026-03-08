package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .thoughts/ directory structure",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	dirs := []string{
		"research", "designs", "structures", "tickets",
		"plans", "specs", "reviews", "archive",
	}
	created := 0
	for _, d := range dirs {
		path := filepath.Join(thoughtsDirFlag, d)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", path)
		created++
	}
	if created == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "already initialized")
	}
	return nil
}
