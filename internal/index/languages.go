package index

import "regexp"

// SymbolPattern maps a regex to a symbol kind. NameGroup is the capture group index for the symbol name.
type SymbolPattern struct {
	Re        *regexp.Regexp
	Kind      string
	NameGroup int
}

// LangConfig defines how to extract symbols from a language.
type LangConfig struct {
	Name       string
	Extensions []string
	Patterns   []SymbolPattern
}

// Languages maps language names to their extraction configs.
var Languages = map[string]*LangConfig{
	"go": {
		Name:       "go",
		Extensions: []string{".go"},
		Patterns: []SymbolPattern{
			{Re: regexp.MustCompile(`^func\s+\([^)]+\)\s+(\w+)\s*\(`), Kind: "method", NameGroup: 1},
			{Re: regexp.MustCompile(`^func\s+(\w+)\s*\(`), Kind: "function", NameGroup: 1},
			{Re: regexp.MustCompile(`^type\s+(\w+)\s+struct\b`), Kind: "struct", NameGroup: 1},
			{Re: regexp.MustCompile(`^type\s+(\w+)\s+interface\b`), Kind: "interface", NameGroup: 1},
			{Re: regexp.MustCompile(`^type\s+(\w+)\s+`), Kind: "type_alias", NameGroup: 1},
		},
	},
	"python": {
		Name:       "python",
		Extensions: []string{".py"},
		Patterns: []SymbolPattern{
			{Re: regexp.MustCompile(`^class\s+(\w+)`), Kind: "class", NameGroup: 1},
			{Re: regexp.MustCompile(`^\s*def\s+(\w+)`), Kind: "function", NameGroup: 1},
		},
	},
	"javascript": {
		Name:       "javascript",
		Extensions: []string{".js", ".jsx"},
		Patterns: []SymbolPattern{
			{Re: regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?function\s+(\w+)`), Kind: "function", NameGroup: 1},
			{Re: regexp.MustCompile(`^(?:export\s+)?class\s+(\w+)`), Kind: "class", NameGroup: 1},
		},
	},
	"typescript": {
		Name:       "typescript",
		Extensions: []string{".ts", ".tsx"},
		Patterns: []SymbolPattern{
			{Re: regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?function\s+(\w+)`), Kind: "function", NameGroup: 1},
			{Re: regexp.MustCompile(`^(?:export\s+)?class\s+(\w+)`), Kind: "class", NameGroup: 1},
			{Re: regexp.MustCompile(`^(?:export\s+)?interface\s+(\w+)`), Kind: "interface", NameGroup: 1},
			{Re: regexp.MustCompile(`^(?:export\s+)?type\s+(\w+)`), Kind: "type_alias", NameGroup: 1},
		},
	},
	"rust": {
		Name:       "rust",
		Extensions: []string{".rs"},
		Patterns: []SymbolPattern{
			{Re: regexp.MustCompile(`^\s+(?:pub(?:\([^)]*\))?\s+)?fn\s+(\w+)`), Kind: "method", NameGroup: 1},
			{Re: regexp.MustCompile(`^(?:pub(?:\([^)]*\))?\s+)?fn\s+(\w+)`), Kind: "function", NameGroup: 1},
			{Re: regexp.MustCompile(`^(?:pub(?:\([^)]*\))?\s+)?struct\s+(\w+)`), Kind: "struct", NameGroup: 1},
			{Re: regexp.MustCompile(`^(?:pub(?:\([^)]*\))?\s+)?enum\s+(\w+)`), Kind: "type_alias", NameGroup: 1},
			{Re: regexp.MustCompile(`^(?:pub(?:\([^)]*\))?\s+)?trait\s+(\w+)`), Kind: "interface", NameGroup: 1},
		},
	},
}

// extensionMap is built at init time for fast extension-based language detection.
var extensionMap map[string]string

func init() {
	extensionMap = make(map[string]string)
	for name, cfg := range Languages {
		for _, ext := range cfg.Extensions {
			extensionMap[ext] = name
		}
	}
}

// DetectLanguage returns the language name for a file path based on extension.
// Returns empty string if the extension is not recognized.
func DetectLanguage(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return extensionMap[path[i:]]
		}
	}
	return ""
}

// GetConfig returns the LangConfig for a language name, or nil if unsupported.
func GetConfig(lang string) *LangConfig {
	return Languages[lang]
}
