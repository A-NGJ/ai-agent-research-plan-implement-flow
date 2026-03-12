package index

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractGo(t *testing.T) {
	dir := t.TempDir()
	src := `package server

func HandleRequest(ctx context.Context) error {
	return nil
}

func (s *Server) Start() error {
	return nil
}

type Server struct {
	addr string
}

type Handler interface {
	Handle()
}

type ID = string
`
	path := writeTestFile(t, dir, "server.go", src)
	cfg := GetConfig("go")

	syms, pkg, err := ExtractSymbols(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkg != "server" {
		t.Errorf("package = %q, want %q", pkg, "server")
	}

	expected := []struct {
		name     string
		kind     string
		exported bool
	}{
		{"HandleRequest", "function", true},
		{"Start", "method", true},
		{"Server", "struct", true},
		{"Handler", "interface", true},
		{"ID", "type_alias", true},
	}

	if len(syms) != len(expected) {
		t.Fatalf("got %d symbols, want %d: %+v", len(syms), len(expected), syms)
	}

	for i, want := range expected {
		if syms[i].Name != want.name {
			t.Errorf("sym[%d].Name = %q, want %q", i, syms[i].Name, want.name)
		}
		if syms[i].Kind != want.kind {
			t.Errorf("sym[%d].Kind = %q, want %q", i, syms[i].Kind, want.kind)
		}
		if syms[i].Exported != want.exported {
			t.Errorf("sym[%d].Exported = %v, want %v", i, syms[i].Exported, want.exported)
		}
	}
}

func TestExtractGoUnexported(t *testing.T) {
	dir := t.TempDir()
	src := `package internal

func helperFunc() {}

type config struct {}
`
	path := writeTestFile(t, dir, "helper.go", src)
	cfg := GetConfig("go")

	syms, _, err := ExtractSymbols(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, s := range syms {
		if s.Exported {
			t.Errorf("symbol %q should not be exported", s.Name)
		}
	}
}

func TestExtractPython(t *testing.T) {
	dir := t.TempDir()
	src := `class UserService:
    def __init__(self):
        pass

    def get_user(self, user_id):
        pass

def create_app():
    pass

def _private_helper():
    pass
`
	path := writeTestFile(t, dir, "service.py", src)
	cfg := GetConfig("python")

	syms, _, err := ExtractSymbols(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []struct {
		name     string
		kind     string
		scope    string
		exported bool
	}{
		{"UserService", "class", "", true},
		{"__init__", "method", "UserService", false},
		{"get_user", "method", "UserService", true},
		{"create_app", "function", "", true},
		{"_private_helper", "function", "", false},
	}

	if len(syms) != len(expected) {
		t.Fatalf("got %d symbols, want %d: %+v", len(syms), len(expected), syms)
	}

	for i, want := range expected {
		if syms[i].Name != want.name {
			t.Errorf("sym[%d].Name = %q, want %q", i, syms[i].Name, want.name)
		}
		if syms[i].Kind != want.kind {
			t.Errorf("sym[%d].Kind = %q, want %q", i, syms[i].Kind, want.kind)
		}
		if syms[i].Scope != want.scope {
			t.Errorf("sym[%d].Scope = %q, want %q", i, syms[i].Scope, want.scope)
		}
		if syms[i].Exported != want.exported {
			t.Errorf("sym[%d].Exported = %v, want %v", i, syms[i].Exported, want.exported)
		}
	}
}

func TestExtractTypeScript(t *testing.T) {
	dir := t.TempDir()
	src := `export interface Config {
  port: number;
}

export type ID = string;

export class Server {
  start() {}
}

export async function main() {}

function helper() {}
`
	path := writeTestFile(t, dir, "server.ts", src)
	cfg := GetConfig("typescript")

	syms, _, err := ExtractSymbols(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []struct {
		name     string
		kind     string
		exported bool
	}{
		{"Config", "interface", true},
		{"ID", "type_alias", true},
		{"Server", "class", true},
		{"main", "function", true},
		{"helper", "function", false},
	}

	if len(syms) != len(expected) {
		t.Fatalf("got %d symbols, want %d: %+v", len(syms), len(expected), syms)
	}

	for i, want := range expected {
		if syms[i].Name != want.name {
			t.Errorf("sym[%d].Name = %q, want %q", i, syms[i].Name, want.name)
		}
		if syms[i].Kind != want.kind {
			t.Errorf("sym[%d].Kind = %q, want %q", i, syms[i].Kind, want.kind)
		}
		if syms[i].Exported != want.exported {
			t.Errorf("sym[%d].Exported = %v, want %v", i, syms[i].Exported, want.exported)
		}
	}
}

func TestExtractJavaScript(t *testing.T) {
	dir := t.TempDir()
	src := `export function fetchData() {}

export class ApiClient {}

function internalHelper() {}
`
	path := writeTestFile(t, dir, "api.js", src)
	cfg := GetConfig("javascript")

	syms, _, err := ExtractSymbols(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(syms) != 3 {
		t.Fatalf("got %d symbols, want 3: %+v", len(syms), syms)
	}
	if syms[0].Name != "fetchData" || !syms[0].Exported {
		t.Errorf("sym[0] = %+v, want fetchData/exported", syms[0])
	}
	if syms[2].Name != "internalHelper" || syms[2].Exported {
		t.Errorf("sym[2] = %+v, want internalHelper/not exported", syms[2])
	}
}

func TestExtractRust(t *testing.T) {
	dir := t.TempDir()
	src := `pub struct Config {
    port: u16,
}

pub trait Handler {
    fn handle(&self);
}

pub enum Status {
    Active,
    Inactive,
}

pub fn create_server() -> Server {}

fn internal_helper() {}

impl Config {
    pub fn new(port: u16) -> Self {}
    fn validate(&self) -> bool {}
}
`
	path := writeTestFile(t, dir, "lib.rs", src)
	cfg := GetConfig("rust")

	syms, _, err := ExtractSymbols(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []struct {
		name     string
		kind     string
		exported bool
	}{
		{"Config", "struct", true},
		{"Handler", "interface", true},
		{"handle", "method", false},
		{"Status", "type_alias", true},
		{"create_server", "function", true},
		{"internal_helper", "function", false},
		{"new", "method", true},
		{"validate", "method", false},
	}

	if len(syms) != len(expected) {
		t.Fatalf("got %d symbols, want %d: %+v", len(syms), len(expected), syms)
	}

	for i, want := range expected {
		if syms[i].Name != want.name {
			t.Errorf("sym[%d].Name = %q, want %q", i, syms[i].Name, want.name)
		}
		if syms[i].Kind != want.kind {
			t.Errorf("sym[%d].Kind = %q, want %q", i, syms[i].Kind, want.kind)
		}
		if syms[i].Exported != want.exported {
			t.Errorf("sym[%d].Exported = %v, want %v", i, syms[i].Exported, want.exported)
		}
	}
}

func TestExtractBinaryFileSkipped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.go")
	// Write a file with null bytes.
	content := []byte("package main\x00\x00func main() {}\x00")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := GetConfig("go")
	syms, _, err := ExtractSymbols(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(syms) != 0 {
		t.Errorf("expected no symbols from binary file, got %d", len(syms))
	}
}

func TestExtractSignature(t *testing.T) {
	dir := t.TempDir()
	src := `package main

func HandleRequest(ctx context.Context, req *Request) (*Response, error) {
}
`
	path := writeTestFile(t, dir, "handler.go", src)
	cfg := GetConfig("go")

	syms, _, err := ExtractSymbols(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(syms) == 0 {
		t.Fatal("expected at least one symbol")
	}
	if syms[0].Signature != "func HandleRequest(ctx context.Context, req *Request) (*Response, error)" {
		t.Errorf("signature = %q", syms[0].Signature)
	}
}

func TestDetectLanguage(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.js", "javascript"},
		{"index.jsx", "javascript"},
		{"server.ts", "typescript"},
		{"component.tsx", "typescript"},
		{"lib.rs", "rust"},
		{"README.md", ""},
		{"Makefile", ""},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := DetectLanguage(tc.path)
			if got != tc.want {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}
