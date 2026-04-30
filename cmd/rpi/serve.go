package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/A-NGJ/rpi/internal/chain"
	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/git"
	"github.com/A-NGJ/rpi/internal/scanner"
	"github.com/A-NGJ/rpi/internal/search"
	tmpl "github.com/A-NGJ/rpi/internal/template"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server exposing all rpi operations as typed tools",
	Long: `Start an MCP (Model Context Protocol) server over stdio.

Exposes all rpi CLI operations as typed MCP tools, enabling AI assistants
to call them with validated schemas instead of constructing shell commands.

Configure your MCP client with:
  {"command": "rpi", "args": ["serve"]}`,
	RunE: runServe,
}

func init() {
	addRpiDirFlag(serveCmd)
	addTemplatesDirFlag(serveCmd)
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	s := newRPIServer()
	return s.Run(context.Background(), &mcp.StdioTransport{})
}

func newRPIServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "rpi",
		Version: "dev",
	}, nil)
	registerTools(s)
	return s
}

// jsonResult marshals v to JSON and returns a TextContent MCP result.
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

type emptyInput struct{}

// mcpDescription derives an MCP tool description from a Cobra command's Long
// and Example fields — single source of truth, no duplicated text.
func mcpDescription(cmd *cobra.Command) string {
	desc := cmd.Long
	if cmd.Example != "" {
		desc += "\n\nExamples:\n" + cmd.Example
	}
	return desc
}

// mcpDescriptionWithPrefix prepends an action-specific one-liner to the parent
// command's description. Used for multi-tool commands (frontmatter, git-context,
// verify, extract) where one Cobra command maps to several MCP tools.
func mcpDescriptionWithPrefix(prefix string, cmd *cobra.Command) string {
	return prefix + "\n\n" + mcpDescription(cmd)
}

func registerTools(s *mcp.Server) {
	// No-param tools — git context (1 Cobra cmd → 3 MCP tools)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_git_context",
		Description: mcpDescriptionWithPrefix("Gather full git context.", gitContextCmd),
	}, handleGitContext)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_git_changed_files",
		Description: mcpDescriptionWithPrefix("List files changed vs main branch.", gitContextCmd),
	}, handleGitChangedFiles)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_git_sensitive_check",
		Description: mcpDescriptionWithPrefix("Scan staged files for sensitive content.", gitContextCmd),
	}, handleGitSensitiveCheck)

	// No-param tools — archive scan
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_archive_scan",
		Description: mcpDescription(archiveScanCmd),
	}, handleArchiveScan)

	// Artifact tools — 1:1 mappings
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_scan",
		Description: mcpDescription(scanCmd),
	}, handleScan)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_scaffold",
		Description: mcpDescription(scaffoldCmd),
	}, handleScaffold)

	// Frontmatter (1 Cobra cmd → 3 MCP tools)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_frontmatter_get",
		Description: mcpDescriptionWithPrefix("Read frontmatter fields from an artifact file.", frontmatterCmd),
	}, handleFrontmatterGet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_frontmatter_set",
		Description: mcpDescriptionWithPrefix("Set a frontmatter field value.", frontmatterCmd),
	}, handleFrontmatterSet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_frontmatter_transition",
		Description: mcpDescriptionWithPrefix("Validated status transition (enforces state machine).", frontmatterCmd),
	}, handleFrontmatterTransition)

	// 1:1 mappings
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_chain",
		Description: mcpDescription(chainCmd),
	}, handleChain)

	// Extract (1 Cobra cmd → 2 MCP tools)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_extract",
		Description: mcpDescriptionWithPrefix("Extract a section from a markdown file.", extractCmd),
	}, handleExtract)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_extract_list_sections",
		Description: mcpDescriptionWithPrefix("List all section headings in a markdown file.", extractCmd),
	}, handleExtractListSections)

	// Verify (1 Cobra cmd → 3 MCP tools)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_verify_completeness",
		Description: mcpDescriptionWithPrefix("Check plan progress: checkbox counts and file coverage.", verifyCmd),
	}, handleVerifyCompleteness)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_verify_markers",
		Description: mcpDescriptionWithPrefix("Scan for TODO/FIXME/HACK markers in source files.", verifyCmd),
	}, handleVerifyMarkers)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_verify_spec",
		Description: mcpDescriptionWithPrefix("Parse spec scenarios into structured JSON with feature name, scenario titles, and Given/When/Then steps.", verifyCmd),
	}, handleVerifySpec)

	// Context
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_context_essentials",
		Description: mcpDescription(contextCmd),
	}, handleContextEssentials)

	// Session awareness
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_session_resume",
		Description: mcpDescription(resumeCmd),
	}, handleSessionResume)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_suggest_next",
		Description: mcpDescription(nextCmd),
	}, handleSuggestNext)

	// Archive (1:1 subcommand mappings)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_archive_check_refs",
		Description: mcpDescription(archiveCheckRefsCmd),
	}, handleArchiveCheckRefs)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_archive_move",
		Description: mcpDescription(archiveMoveCmd),
	}, handleArchiveMove)

	// Semantic search via optional qmd backend
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_search",
		Description: mcpDescription(searchCmd),
	}, handleSearch)
}

// --- No-param tool handlers ---

func handleGitContext(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	result, err := git.GatherContext()
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleGitChangedFiles(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	files, err := git.ChangedFiles()
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(files)
}

func handleGitSensitiveCheck(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	matches, err := git.SensitiveCheck()
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(matches)
}

func handleArchiveScan(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	results, err := scanner.Scan(rpiDirFlag, scanner.Filters{Archivable: true})
	if err != nil {
		return nil, nil, err
	}

	var output []archiveScanResult
	for _, r := range results {
		refPath := r.Path
		if rel, relErr := filepath.Rel(rpiDirFlag, r.Path); relErr == nil {
			refPath = rel
		}
		refCount, _ := scanner.CountReferences(rpiDirFlag, refPath)
		output = append(output, archiveScanResult{
			Path:     r.Path,
			Type:     r.Type,
			Status:   r.Status,
			Title:    r.Title,
			RefCount: refCount,
		})
	}
	if output == nil {
		output = []archiveScanResult{}
	}
	return jsonResult(output)
}

// --- Parameterized tool input structs ---

type scanInput struct {
	Type       string `json:"type,omitempty" jsonschema:"filter by artifact type (plan, design, research, spec, review)"`
	Status     string `json:"status,omitempty" jsonschema:"filter by status (draft, active, complete, superseded)"`
	Design     string `json:"design,omitempty" jsonschema:"filter by frontmatter design field"`
	References string `json:"references,omitempty" jsonschema:"find files that reference this path"`
	Archivable bool   `json:"archivable,omitempty" jsonschema:"show only archivable artifacts"`
}

type scaffoldInput struct {
	Type     string `json:"type" jsonschema:"artifact type: research, design, diagnosis, plan, verify-report, spec"`
	Topic    string `json:"topic" jsonschema:"topic or title for the artifact"`
	Ticket   string `json:"ticket,omitempty" jsonschema:"ticket ID"`
	Research string `json:"research,omitempty" jsonschema:"path to research document"`
	Design   string `json:"design,omitempty" jsonschema:"path to design document"`
	Spec     string `json:"spec,omitempty" jsonschema:"path to spec document"`
	Tags     string `json:"tags,omitempty" jsonschema:"comma-separated tags"`
	Write    bool   `json:"write,omitempty" jsonschema:"write to file instead of returning content"`
	Force    bool   `json:"force,omitempty" jsonschema:"allow overwriting existing files"`
}

type frontmatterGetInput struct {
	File  string `json:"file" jsonschema:"path to the artifact file"`
	Field string `json:"field,omitempty" jsonschema:"specific field to read (omit for all frontmatter)"`
}

type frontmatterSetInput struct {
	File  string `json:"file" jsonschema:"path to the artifact file"`
	Field string `json:"field" jsonschema:"frontmatter field name"`
	Value string `json:"value" jsonschema:"value to set"`
}

type frontmatterTransitionInput struct {
	File   string `json:"file" jsonschema:"path to the artifact file"`
	Status string `json:"status" jsonschema:"target status (draft, active, complete, superseded, archived)"`
}

type chainInput struct {
	Path     string `json:"path" jsonschema:"path to the root artifact"`
	Sections string `json:"sections,omitempty" jsonschema:"comma-separated section names to extract from each artifact"`
}

type extractInput struct {
	Path    string `json:"path" jsonschema:"path to the markdown file"`
	Section string `json:"section" jsonschema:"section heading to extract (case-insensitive prefix match)"`
}

type extractListSectionsInput struct {
	Path string `json:"path" jsonschema:"path to the markdown file"`
}

type verifyCompletenessInput struct {
	PlanPath string `json:"plan_path" jsonschema:"path to the plan file"`
}

type verifyMarkersInput struct {
	FilePath string `json:"file_path,omitempty" jsonschema:"specific file to scan (omit to scan git-changed files)"`
}

type verifySpecInput struct {
	SpecPath string `json:"spec_path" jsonschema:"path to the spec file"`
}

type contextInput struct {
	PlanPath string `json:"plan_path,omitempty" jsonschema:"path to a specific plan file (omit to auto-detect the most recent active plan)"`
}

type suggestInput struct {
	ArtifactPath string `json:"artifact_path,omitempty" jsonschema:"path to a specific artifact (omit to auto-detect from all artifacts)"`
}

type archiveCheckRefsInput struct {
	Path string `json:"path" jsonschema:"path to check for references"`
}

type archiveMoveInput struct {
	Path  string `json:"path" jsonschema:"path to the artifact to archive"`
	Force bool   `json:"force,omitempty" jsonschema:"skip active reference check"`
}

type searchInput struct {
	Query          string  `json:"query" jsonschema:"natural-language search query"`
	Type           string  `json:"type,omitempty" jsonschema:"filter by artifact type (research, design, plan, spec, diagnosis, review)"`
	Limit          int     `json:"limit,omitempty" jsonschema:"maximum hits to return (default 5, max 20)"`
	IncludeArchive bool    `json:"include_archive,omitempty" jsonschema:"include archived artifacts in results"`
	MinScore       float64 `json:"min_score,omitempty" jsonschema:"minimum relevance score (0.0-1.0)"`
}

// --- Parameterized tool handlers ---

func handleScan(_ context.Context, _ *mcp.CallToolRequest, input scanInput) (*mcp.CallToolResult, any, error) {
	results, err := scanner.Scan(rpiDirFlag, scanner.Filters{
		Status:     input.Status,
		Type:       input.Type,
		Design:     input.Design,
		References: input.References,
		Archivable: input.Archivable,
	})
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(results)
}

func handleScaffold(_ context.Context, _ *mcp.CallToolRequest, input scaffoldInput) (*mcp.CallToolResult, any, error) {
	if _, ok := typeDirs[input.Type]; !ok {
		return nil, nil, fmt.Errorf("unknown artifact type %q; valid types: %v", input.Type, validTypes)
	}
	if input.Topic == "" {
		return nil, nil, fmt.Errorf("topic is required")
	}

	ctx := &tmpl.RenderContext{
		Type:     input.Type,
		Topic:    input.Topic,
		Ticket:   input.Ticket,
		Research: input.Research,
		Design:   input.Design,
		Spec:     input.Spec,
		Tags:     input.Tags,
	}

	labels := map[string]string{
		"research": "Research", "plan": "Plan", "design": "Design",
		"diagnosis": "Diagnosis", "verify-report": "Verification Report", "spec": "Spec",
	}
	ctx.TypeLabel = labels[input.Type]

	if err := tmpl.ResolveAutoVars(ctx); err != nil {
		return nil, nil, err
	}
	ctx.Filename = tmpl.GenerateFilename(input.Type, ctx)

	output, err := tmpl.RenderTemplate(input.Type, ctx, resolveTemplatesDir())
	if err != nil {
		return nil, nil, err
	}

	if !input.Write {
		return jsonResult(map[string]string{"content": output, "filename": ctx.Filename})
	}

	subdir := typeDirs[input.Type]
	dir := filepath.Join(rpiDirFlag, subdir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, err
	}
	outPath := filepath.Join(dir, ctx.Filename)
	if !input.Force {
		if _, err := os.Stat(outPath); err == nil {
			return nil, nil, fmt.Errorf("file already exists: %s (use force=true to overwrite)", outPath)
		}
	}
	if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]string{"path": outPath})
}

func handleFrontmatterGet(_ context.Context, _ *mcp.CallToolRequest, input frontmatterGetInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.File)
	if err != nil {
		return nil, nil, err
	}
	if input.Field != "" {
		val, ok := doc.Frontmatter[input.Field]
		if !ok {
			return jsonResult(nil)
		}
		return jsonResult(val)
	}
	return jsonResult(doc.Frontmatter)
}

func handleFrontmatterSet(_ context.Context, _ *mcp.CallToolRequest, input frontmatterSetInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.File)
	if err != nil {
		return nil, nil, err
	}
	doc.Frontmatter[input.Field] = input.Value
	if err := frontmatter.Write(doc); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]string{"status": "ok", "file": input.File, "field": input.Field, "value": input.Value})
}

func handleFrontmatterTransition(_ context.Context, _ *mcp.CallToolRequest, input frontmatterTransitionInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.File)
	if err != nil {
		return nil, nil, err
	}
	if err := frontmatter.Transition(doc, input.Status); err != nil {
		return nil, nil, err
	}
	if err := frontmatter.Write(doc); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]string{"status": "ok", "file": input.File, "new_status": input.Status})
}

func handleChain(_ context.Context, _ *mcp.CallToolRequest, input chainInput) (*mcp.CallToolResult, any, error) {
	opts := chain.ResolveOptions{}
	if input.Sections != "" {
		parts := strings.Split(input.Sections, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		opts.Sections = parts
	}
	result, err := chain.Resolve(input.Path, opts)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleExtract(_ context.Context, _ *mcp.CallToolRequest, input extractInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.Path)
	if err != nil {
		return nil, nil, err
	}
	content, ok := frontmatter.ExtractSection(doc.Body, input.Section)
	if !ok {
		return nil, nil, fmt.Errorf("section %q not found", input.Section)
	}
	return jsonResult(map[string]string{
		"path": input.Path, "section": input.Section, "content": content,
	})
}

func handleExtractListSections(_ context.Context, _ *mcp.CallToolRequest, input extractListSectionsInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.Path)
	if err != nil {
		return nil, nil, err
	}
	sections := frontmatter.ListSections(doc.Body)
	return jsonResult(map[string]any{
		"path": input.Path, "sections": sections,
	})
}

func handleVerifyCompleteness(_ context.Context, _ *mcp.CallToolRequest, input verifyCompletenessInput) (*mcp.CallToolResult, any, error) {
	content, err := os.ReadFile(input.PlanPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading plan: %w", err)
	}
	checkboxes := parseCheckboxes(string(content))
	planFiles := extractPlanFiles(string(content))
	gitFiles, err := git.ChangedFiles()
	if err != nil {
		gitFiles = []string{}
	}
	compare := comparePlanVsGit(planFiles, gitFiles)
	result := CompletenessResult{
		CheckboxResult: checkboxes,
		CompareResult:  compare,
	}
	return jsonResult(result)
}

func handleVerifyMarkers(_ context.Context, _ *mcp.CallToolRequest, input verifyMarkersInput) (*mcp.CallToolResult, any, error) {
	files, err := filesToScan(input.FilePath)
	if err != nil {
		return nil, nil, err
	}
	result := MarkersResult{
		Markers: []Marker{},
		Count:   map[string]int{"TODO": 0, "FIXME": 0, "HACK": 0},
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		markers := scanMarkers(file, string(data))
		result.Markers = append(result.Markers, markers...)
		for _, m := range markers {
			result.Count[m.Type]++
		}
	}
	return jsonResult(result)
}

func handleVerifySpec(_ context.Context, _ *mcp.CallToolRequest, input verifySpecInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.SpecPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading spec: %w", err)
	}
	feature, _ := doc.Frontmatter["feature"].(string)
	scenarios := parseScenarios(doc.Body)
	result := SpecResult{
		Spec:      input.SpecPath,
		Feature:   feature,
		Scenarios: scenarios,
		Total:     len(scenarios),
	}
	return jsonResult(result)
}

func handleContextEssentials(_ context.Context, _ *mcp.CallToolRequest, input contextInput) (*mcp.CallToolResult, any, error) {
	result, err := assembleContext(rpiDirFlag, input.PlanPath)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleSessionResume(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	result, err := assembleResume(rpiDirFlag)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleSuggestNext(_ context.Context, _ *mcp.CallToolRequest, input suggestInput) (*mcp.CallToolResult, any, error) {
	result, err := suggestNext(rpiDirFlag, input.ArtifactPath)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleArchiveCheckRefs(_ context.Context, _ *mcp.CallToolRequest, input archiveCheckRefsInput) (*mcp.CallToolResult, any, error) {
	refs, err := scanner.FindReferences(rpiDirFlag, input.Path)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(refs)
}

func handleArchiveMove(_ context.Context, _ *mcp.CallToolRequest, input archiveMoveInput) (*mcp.CallToolResult, any, error) {
	result, err := doArchiveMove(input.Path, rpiDirFlag, input.Force, time.Now())
	if err == errHasReferences {
		return nil, nil, fmt.Errorf("file has active references (use force=true to override)")
	}
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleSearch(ctx context.Context, _ *mcp.CallToolRequest, input searchInput) (*mcp.CallToolResult, any, error) {
	resp := search.Query(ctx, rpiDirFlag, search.SearchParams{
		Query:          input.Query,
		Type:           input.Type,
		Limit:          input.Limit,
		IncludeArchive: input.IncludeArchive,
		MinScore:       input.MinScore,
	}, search.QueryOptions{})
	return jsonResult(resp)
}
