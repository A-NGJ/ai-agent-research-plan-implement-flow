package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
	"github.com/spf13/cobra"
)

var (
	sectionFlag      string
	listSectionsFlag bool
)

var extractCmd = &cobra.Command{
	Use:   "extract <path>",
	Short: "Extract a section from a markdown file",
	Long:  "Parse a markdown file and extract the content of a specific ## heading by case-insensitive prefix match, or list all available section headings.",
	Args:  cobra.ExactArgs(1),
	RunE:  runExtract,
}

func init() {
	addFormatFlag(extractCmd)
	extractCmd.Flags().StringVar(&sectionFlag, "section", "", "Section heading to extract")
	extractCmd.Flags().BoolVar(&listSectionsFlag, "list-sections", false, "List all ## section headings in the file")
	rootCmd.AddCommand(extractCmd)
}

func runExtract(cmd *cobra.Command, args []string) error {
	if sectionFlag != "" && listSectionsFlag {
		return fmt.Errorf("cannot use --section and --list-sections together")
	}
	if sectionFlag == "" && !listSectionsFlag {
		return fmt.Errorf("either --section or --list-sections is required")
	}

	doc, err := frontmatter.Parse(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if listSectionsFlag {
		return runListSections(doc, args[0])
	}

	return runExtractSection(doc, args[0])
}

func runListSections(doc *frontmatter.Document, path string) error {
	sections := frontmatter.ListSections(doc.Body)
	if len(sections) == 0 {
		fmt.Fprintln(os.Stderr, "no sections found")
		os.Exit(1)
	}

	format := formatFlag
	if format == "" {
		format = "text"
	}

	switch format {
	case "text":
		for _, s := range sections {
			// Strip "## " prefix for clean text output
			fmt.Println(strings.TrimPrefix(s, "## "))
		}
	case "json":
		out := map[string]interface{}{
			"path":     path,
			"sections": sections,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
	case "md":
		for _, s := range sections {
			fmt.Printf("- %s\n", s)
		}
	default:
		return fmt.Errorf("unknown format: %s (expected json, md, or text)", format)
	}

	return nil
}

func runExtractSection(doc *frontmatter.Document, path string) error {
	content, ok := frontmatter.ExtractSection(doc.Body, sectionFlag)
	if !ok {
		fmt.Fprintf(os.Stderr, "section %q not found\n", sectionFlag)
		os.Exit(1)
	}

	format := formatFlag
	if format == "" {
		format = "text"
	}

	switch format {
	case "text", "md":
		fmt.Print(content)
	case "json":
		out := map[string]string{
			"path":    path,
			"section": sectionFlag,
			"content": content,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s (expected json, md, or text)", format)
	}

	return nil
}
