package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupBootstrapEnv configures HOME, optionally seeds a global install for
// the given target, optionally creates a git root at projectDir, and
// chdirs cwd into projectDir (or a subdir of it). Each piece is
// independently configurable so tests can exercise the different no-op
// paths.
type bootstrapSetup struct {
	home              string
	globalTarget      string // "claude", "opencode", "" (none), "both"
	projectDir        string
	createGit         bool
	createPreExisting bool   // pre-create projectDir/.rpi/ to test idempotent path
	cwdSubdir         string // relative to projectDir; "" means cwd = projectDir
}

func setupBootstrap(t *testing.T, s bootstrapSetup) string {
	t.Helper()
	t.Setenv("HOME", s.home)

	if s.globalTarget == "claude" || s.globalTarget == "both" {
		// Seed the marker subdir bootstrap probes for.
		if err := os.MkdirAll(filepath.Join(s.home, ".claude", "skills", "rpi-research"), 0755); err != nil {
			t.Fatalf("seed claude global: %v", err)
		}
	}
	if s.globalTarget == "opencode" || s.globalTarget == "both" {
		if err := os.MkdirAll(filepath.Join(s.home, ".config", "opencode", "skills", "rpi-research"), 0755); err != nil {
			t.Fatalf("seed opencode global: %v", err)
		}
	}

	if s.createGit {
		if err := os.MkdirAll(filepath.Join(s.projectDir, ".git"), 0755); err != nil {
			t.Fatalf("create .git/: %v", err)
		}
	}
	if s.createPreExisting {
		if err := os.MkdirAll(filepath.Join(s.projectDir, ".rpi"), 0755); err != nil {
			t.Fatalf("pre-create .rpi/: %v", err)
		}
	}

	cwd := s.projectDir
	if s.cwdSubdir != "" {
		cwd = filepath.Join(s.projectDir, s.cwdSubdir)
		if err := os.MkdirAll(cwd, 0755); err != nil {
			t.Fatalf("create cwd subdir: %v", err)
		}
	}
	t.Chdir(cwd)
	return cwd
}

func runBootstrapCmd(t *testing.T) (stdout, stderr *bytes.Buffer, err error) {
	t.Helper()
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	cmd := bootstrapCmd
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	err = cmd.RunE(cmd, nil)
	return
}

func TestBootstrapNoOpWhenAlreadyInitialized(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	setupBootstrap(t, bootstrapSetup{
		home:              home,
		globalTarget:      "claude",
		projectDir:        projectDir,
		createGit:         true,
		createPreExisting: true,
	})

	_, stderr, err := runBootstrapCmd(t)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected silent no-op, got stderr: %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(projectDir, "CLAUDE.md")); err == nil {
		t.Error("rules file should not be written when .rpi/ already exists")
	}
}

func TestBootstrapNoOpWhenNoGlobalInstall(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	setupBootstrap(t, bootstrapSetup{
		home:         home,
		globalTarget: "",
		projectDir:   projectDir,
		createGit:    true,
	})

	_, stderr, err := runBootstrapCmd(t)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected silent no-op, got stderr: %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".rpi")); err == nil {
		t.Error(".rpi/ should not be created without a global install")
	}
}

func TestBootstrapNoOpOutsideGit(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	setupBootstrap(t, bootstrapSetup{
		home:         home,
		globalTarget: "claude",
		projectDir:   projectDir,
		createGit:    false,
	})

	_, stderr, err := runBootstrapCmd(t)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected silent no-op, got stderr: %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".rpi")); err == nil {
		t.Error(".rpi/ should not be created outside a git repo")
	}
}

func TestBootstrapAutoInitsAtGitRoot(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	cwd := setupBootstrap(t, bootstrapSetup{
		home:         home,
		globalTarget: "claude",
		projectDir:   projectDir,
		createGit:    true,
		cwdSubdir:    "sub/dir",
	})
	_ = cwd

	_, stderr, err := runBootstrapCmd(t)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	// Auto-init notice on stderr.
	if !strings.Contains(stderr.String(), "Auto-initialized .rpi/") {
		t.Errorf("expected auto-init notice, got stderr: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "skills inherited from") {
		t.Errorf("notice should mention global path, got: %q", stderr.String())
	}

	// Project-side artifacts at the git root.
	for _, want := range []string{".rpi/plans", ".rpi/specs", ".rpi/templates", "CLAUDE.md", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(projectDir, want)); err != nil {
			t.Errorf("missing %s under git root: %v", want, err)
		}
	}

	gitignore, _ := os.ReadFile(filepath.Join(projectDir, ".gitignore"))
	if !strings.Contains(string(gitignore), ".rpi/*") || !strings.Contains(string(gitignore), "!.rpi/specs/") {
		t.Errorf(".gitignore missing standard policy: %s", gitignore)
	}

	// Bootstrap must not duplicate the global install at the project level.
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err == nil {
		t.Error("bootstrap should not create .claude/ at project root")
	}
}

func TestBootstrapIdempotentAfterAutoInit(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	setupBootstrap(t, bootstrapSetup{
		home:         home,
		globalTarget: "claude",
		projectDir:   projectDir,
		createGit:    true,
	})

	if _, _, err := runBootstrapCmd(t); err != nil {
		t.Fatalf("first bootstrap: %v", err)
	}
	_, stderr2, err := runBootstrapCmd(t)
	if err != nil {
		t.Fatalf("second bootstrap: %v", err)
	}
	if stderr2.Len() != 0 {
		t.Errorf("second run should be silent, got: %q", stderr2.String())
	}
}

func TestBootstrapPrefersClaudeOverOpenCode(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	setupBootstrap(t, bootstrapSetup{
		home:         home,
		globalTarget: "both",
		projectDir:   projectDir,
		createGit:    true,
	})

	_, stderr, err := runBootstrapCmd(t)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if !strings.Contains(stderr.String(), filepath.Join(home, ".claude")) {
		t.Errorf("expected Claude global path in notice when both are present, got: %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(projectDir, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md should be installed when Claude target wins")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "AGENTS.md")); err == nil {
		t.Error("AGENTS.md should not be installed when Claude target wins")
	}
}

func TestBootstrap_AppendsMissingSectionsToExistingRulesFile(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	setupBootstrap(t, bootstrapSetup{
		home:         home,
		globalTarget: "claude",
		projectDir:   projectDir,
		createGit:    true,
	})

	// User has a hand-rolled CLAUDE.md before bootstrap runs.
	claudeMD := filepath.Join(projectDir, "CLAUDE.md")
	prior := "# CLAUDE.md\n\n## Project Overview\n\nUser-owned overview text.\n"
	if err := os.WriteFile(claudeMD, []byte(prior), 0644); err != nil {
		t.Fatalf("seed CLAUDE.md: %v", err)
	}

	if _, _, err := runBootstrapCmd(t); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	got, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	// User content preserved at the start of the file.
	if !bytes.HasPrefix(got, []byte(prior)) {
		t.Errorf("user content not preserved at file start.\nwant prefix: %q\ngot:         %q", prior, got)
	}
	// Missing template sections appended.
	for _, want := range []string{"## Git Workflow", "## RPI Artifacts Directory", "## Codebase Navigation", "## Development Conventions"} {
		if !bytes.Contains(got, []byte("\n"+want+"\n")) {
			t.Errorf("missing template section %q not appended by bootstrap.\nfile:\n%s", want, got)
		}
	}
	// Contract block appended too.
	if !bytes.Contains(got, []byte("<!-- rpi:contract:begin")) {
		t.Error("bootstrap did not append contract block to existing CLAUDE.md")
	}
}

func TestBootstrapDetectsOpenCodeAlone(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	setupBootstrap(t, bootstrapSetup{
		home:         home,
		globalTarget: "opencode",
		projectDir:   projectDir,
		createGit:    true,
	})

	_, stderr, err := runBootstrapCmd(t)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if !strings.Contains(stderr.String(), filepath.Join(".config", "opencode")) {
		t.Errorf("expected opencode global path in notice, got: %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(projectDir, "AGENTS.md")); err != nil {
		t.Error("AGENTS.md should be installed for opencode target")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md should not be installed for opencode target")
	}
}
