package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
	"github.com/spf13/cobra"
)

var frontmatterCmd = &cobra.Command{
	Use:   "frontmatter <action> <file> [args...]",
	Short: "Read, modify, and validate YAML frontmatter in .thoughts/ files",
	Long: `Actions:
  get <file> [field]         Read a field or all frontmatter as JSON
  set <file> <field> <value> Set a frontmatter field
  transition <file> <status> Validated status transition`,
	Args: cobra.MinimumNArgs(2),
	RunE: runFrontmatter,
}

func init() {
	rootCmd.AddCommand(frontmatterCmd)
}

func runFrontmatter(cmd *cobra.Command, args []string) error {
	action := args[0]
	file := args[1]

	switch action {
	case "get":
		return runGet(file, args[2:])
	case "set":
		if len(args) < 4 {
			return fmt.Errorf("usage: rpi frontmatter set <file> <field> <value>")
		}
		return runSet(file, args[2], args[3])
	case "transition":
		if len(args) < 3 {
			return fmt.Errorf("usage: rpi frontmatter transition <file> <status>")
		}
		return runTransition(file, args[2])
	default:
		return fmt.Errorf("unknown action: %s (expected get, set, or transition)", action)
	}
}

func runGet(file string, args []string) error {
	doc, err := frontmatter.Parse(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		val, ok := doc.Frontmatter[args[0]]
		if !ok {
			fmt.Println("null")
			return nil
		}
		data, _ := json.Marshal(val)
		fmt.Println(string(data))
		return nil
	}

	data, _ := json.MarshalIndent(doc.Frontmatter, "", "  ")
	fmt.Println(string(data))
	return nil
}

func runSet(file, field, value string) error {
	doc, err := frontmatter.Parse(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	doc.Frontmatter[field] = value
	if err := frontmatter.Write(doc); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return nil
}

func runTransition(file, status string) error {
	doc, err := frontmatter.Parse(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := frontmatter.Transition(doc, status); err != nil {
		var ve *frontmatter.ValidationError
		if errors.As(err, &ve) {
			fmt.Fprintf(os.Stderr, "error: %v\n", ve)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := frontmatter.Write(doc); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return nil
}
