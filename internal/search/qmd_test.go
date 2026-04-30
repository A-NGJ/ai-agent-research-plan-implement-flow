package search

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// stubRunner builds a runner that returns canned (output, err) per command name.
func stubRunner(t *testing.T, responses map[string]stubResponse) runner {
	t.Helper()
	return func(_ context.Context, name string, args ...string) ([]byte, error) {
		key := name
		if len(args) > 0 {
			key = name + " " + args[0]
		}
		r, ok := responses[key]
		if !ok {
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		return []byte(r.out), r.err
	}
}

type stubResponse struct {
	out string
	err error
}

func TestIsAvailable(t *testing.T) {
	t.Run("qmd on path returns true and caches", func(t *testing.T) {
		resetAvailabilityCache()
		t.Cleanup(resetAvailabilityCache)

		callCount := 0
		c := NewClient().WithRunner(func(_ context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			return []byte("qmd 1.0.0"), nil
		})

		if !c.IsAvailable(context.Background()) {
			t.Fatal("expected IsAvailable to be true")
		}
		// Second call: cached, runner not re-invoked.
		if !c.IsAvailable(context.Background()) {
			t.Fatal("expected IsAvailable to be true on second call")
		}
		if callCount != 1 {
			t.Fatalf("expected runner to be called once (cached), got %d", callCount)
		}
	})

	t.Run("qmd missing returns false", func(t *testing.T) {
		resetAvailabilityCache()
		t.Cleanup(resetAvailabilityCache)

		c := NewClient().WithRunner(func(_ context.Context, name string, args ...string) ([]byte, error) {
			return nil, errors.New("exec: \"qmd\": executable file not found in $PATH")
		})

		if c.IsAvailable(context.Background()) {
			t.Fatal("expected IsAvailable to be false when qmd missing")
		}
	})
}

func TestStatusParser(t *testing.T) {
	// Isolate from any host filesystem cache — these tests assert solely on
	// the parsed qmd-status output, with the on-disk fallback neutralized.
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	cases := []struct {
		name              string
		statusOut         string
		statusErr         error
		isAvailable       bool
		wantInstalled     bool
		wantDaemonRunning bool
		wantModelsReady   bool
		wantErr           bool
	}{
		{
			name:              "mcp running with models ready",
			statusOut:         "MCP: running (PID 123)\nModels: ready\n",
			isAvailable:       true,
			wantInstalled:     true,
			wantDaemonRunning: true,
			wantModelsReady:   true,
		},
		{
			name:              "mcp not running, models ready",
			statusOut:         "MCP: stopped\nModels: ready\n",
			isAvailable:       true,
			wantInstalled:     true,
			wantDaemonRunning: false,
			wantModelsReady:   true,
		},
		{
			name:              "mcp running, models not present",
			statusOut:         "MCP: running\n",
			isAvailable:       true,
			wantInstalled:     true,
			wantDaemonRunning: true,
			wantModelsReady:   false,
		},
		{
			name:          "qmd missing",
			isAvailable:   false,
			wantInstalled: false,
		},
		{
			name:              "malformed status output",
			statusOut:         "garbage data with no recognizable markers",
			isAvailable:       true,
			wantInstalled:     true,
			wantDaemonRunning: false,
			wantModelsReady:   false,
		},
		{
			name:          "qmd status returns error",
			statusOut:     "partial output",
			statusErr:     errors.New("qmd status: exit 1"),
			isAvailable:   true,
			wantInstalled: true,
			wantErr:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetAvailabilityCache()
			t.Cleanup(resetAvailabilityCache)

			run := func(_ context.Context, name string, args ...string) ([]byte, error) {
				if len(args) > 0 && args[0] == "--version" {
					if tc.isAvailable {
						return []byte("qmd 1.0.0"), nil
					}
					return nil, errors.New("not found")
				}
				if len(args) > 0 && args[0] == "status" {
					return []byte(tc.statusOut), tc.statusErr
				}
				t.Fatalf("unexpected command: %s %v", name, args)
				return nil, nil
			}

			c := NewClient().WithRunner(run)
			state, err := c.Status(context.Background())

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error from Status, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if state.Installed != tc.wantInstalled {
				t.Errorf("Installed: got %v, want %v", state.Installed, tc.wantInstalled)
			}
			if state.DaemonRunning != tc.wantDaemonRunning {
				t.Errorf("DaemonRunning: got %v, want %v", state.DaemonRunning, tc.wantDaemonRunning)
			}
			if state.ModelsReady != tc.wantModelsReady {
				t.Errorf("ModelsReady: got %v, want %v", state.ModelsReady, tc.wantModelsReady)
			}
		})
	}
}

func TestModelsCachedOnDisk(t *testing.T) {
	t.Run("returns true when gguf file present", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", dir)
		modelsDir := filepath.Join(dir, "qmd", "models")
		if err := os.MkdirAll(modelsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(modelsDir, "embed.gguf"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if !modelsCachedOnDisk() {
			t.Fatal("expected true with gguf file present")
		}
	})

	t.Run("returns false when models dir empty", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", dir)
		if err := os.MkdirAll(filepath.Join(dir, "qmd", "models"), 0o755); err != nil {
			t.Fatal(err)
		}
		if modelsCachedOnDisk() {
			t.Fatal("expected false when models dir empty")
		}
	})

	t.Run("returns false when models dir missing", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", dir)
		if modelsCachedOnDisk() {
			t.Fatal("expected false when models dir missing")
		}
	})
}
