package index

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	goPackageRe     = regexp.MustCompile(`^package\s+(\w+)`)
	pyClassIndentRe = regexp.MustCompile(`^class\s+`)
)

// ExtractSymbols scans a file line-by-line and returns all matched symbols plus the detected package name.
func ExtractSymbols(filePath string, cfg *LangConfig) ([]Symbol, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	if isBinary(f) {
		return nil, "", nil
	}
	// Reset after binary check.
	if _, err := f.Seek(0, 0); err != nil {
		return nil, "", err
	}

	lang := cfg.Name
	var pkg string
	var symbols []Symbol
	var currentScope string

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Package detection.
		if pkg == "" {
			pkg = detectPackage(lang, line, filePath)
		}

		// Scope tracking.
		currentScope = updateScope(lang, line, currentScope)

		// Pattern matching — first match wins.
		for _, pat := range cfg.Patterns {
			m := pat.Re.FindStringSubmatch(line)
			if m == nil || pat.NameGroup >= len(m) {
				continue
			}
			name := m[pat.NameGroup]
			kind := pat.Kind

			// For Python, indented def inside a class is a method.
			if lang == "python" && kind == "function" && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
				kind = "method"
			}

			sym := Symbol{
				Name:      name,
				Kind:      kind,
				File:      filePath,
				Line:      lineNum,
				Package:   pkg,
				Scope:     scopeFor(lang, kind, currentScope),
				Signature: trimSignature(line),
				Exported:  isExported(lang, name, line),
			}
			symbols = append(symbols, sym)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return symbols, pkg, nil
}

func detectPackage(lang, line, filePath string) string {
	switch lang {
	case "go":
		if m := goPackageRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	case "python":
		return filepath.Base(filepath.Dir(filePath))
	case "javascript", "typescript":
		return filepath.Base(filepath.Dir(filePath))
	case "rust":
		return filepath.Base(filepath.Dir(filePath))
	}
	return ""
}

func updateScope(lang, line, current string) string {
	switch lang {
	case "go":
		// Go methods have receivers — scope is tracked per-symbol via receiver syntax, not via block nesting.
		// Structs/interfaces don't nest functions, so no scope tracking needed.
		return ""
	case "python":
		// Top-level class declaration resets scope.
		if pyClassIndentRe.MatchString(line) {
			m := Languages["python"].Patterns[0].Re.FindStringSubmatch(line)
			if m != nil {
				return m[1]
			}
		}
		// Non-indented non-class line resets scope.
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '#' {
			return ""
		}
		return current
	case "javascript", "typescript":
		for _, pat := range Languages[lang].Patterns {
			if pat.Kind == "class" {
				if m := pat.Re.FindStringSubmatch(line); m != nil {
					return m[1]
				}
			}
		}
		// Closing brace at column 0 resets scope.
		if strings.TrimSpace(line) == "}" {
			return ""
		}
		return current
	case "rust":
		// impl block tracking.
		implRe := regexp.MustCompile(`^impl\s+(?:\w+\s+for\s+)?(\w+)`)
		if m := implRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
		if strings.TrimSpace(line) == "}" && current != "" {
			return ""
		}
		return current
	}
	return current
}

func scopeFor(lang, kind, currentScope string) string {
	switch lang {
	case "go":
		// Go scope is extracted from method receiver, not from block nesting.
		return ""
	default:
		if kind == "method" {
			return currentScope
		}
		return ""
	}
}

func trimSignature(line string) string {
	s := strings.TrimSpace(line)
	// Strip trailing opening brace.
	s = strings.TrimRight(s, " {")
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}

func isExported(lang, name, line string) bool {
	switch lang {
	case "go":
		if len(name) == 0 {
			return false
		}
		return unicode.IsUpper(rune(name[0]))
	case "python":
		return !strings.HasPrefix(name, "_")
	case "javascript", "typescript":
		return strings.HasPrefix(strings.TrimSpace(line), "export")
	case "rust":
		trimmed := strings.TrimSpace(line)
		return strings.HasPrefix(trimmed, "pub ") || strings.HasPrefix(trimmed, "pub(")
	}
	return true
}

// isBinary checks the first 512 bytes for null bytes.
func isBinary(f *os.File) bool {
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil || n == 0 {
		return false
	}
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}
