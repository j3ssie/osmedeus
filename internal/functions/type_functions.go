package functions

import (
	"bufio"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// Type detection constants
const (
	TypeFile   = "file"
	TypeFolder = "folder"
	TypeCIDR   = "cidr"
	TypeIP     = "ip"
	TypeURL    = "url"
	TypeDomain = "domain"
	TypeString = "string"
)

// Compiled regex patterns for type detection
var (
	// CIDR patterns (IPv4 and IPv6)
	cidrV4Pattern = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)
	cidrV6Pattern = regexp.MustCompile(`^([0-9a-fA-F:]+)/\d{1,3}$`)

	// Domain pattern - matches valid domain names
	domainPattern = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
)

// getTypes detects the type of the input string
// Returns: "file", "folder", "cidr", "ip", "url", "domain", or "string"
// Detection order (most specific first):
// 1. file - os.Stat() succeeds and is not a directory
// 2. folder - os.Stat() succeeds and is a directory
// 3. cidr - matches CIDR pattern and validates with net.ParseCIDR()
// 4. ip - validates with net.ParseIP()
// 5. url - has http:// or https:// scheme
// 6. domain - matches domain pattern regex
// 7. string - fallback for anything else
func (vf *vmFunc) getTypes(call goja.FunctionCall) goja.Value {
	input := call.Argument(0).String()
	log := logger.Get()

	log.Debug("Calling "+terminal.HiGreen("get_types"), zap.String("input", input))

	if input == "undefined" || input == "" {
		log.Debug(terminal.HiGreen("get_types")+" result", zap.String("type", TypeString))
		return vf.vm.ToValue(TypeString)
	}

	result := detectInputType(input)
	log.Debug(terminal.HiGreen("get_types")+" result", zap.String("input", input), zap.String("type", result))

	return vf.vm.ToValue(result)
}

// detectInputType determines the type of the given input
func detectInputType(input string) string {
	// 1. Check for file or folder (most specific - actual filesystem check)
	if info, err := os.Stat(input); err == nil {
		if info.IsDir() {
			return TypeFolder
		}
		return TypeFile
	}

	// 2. Check for CIDR notation
	if isCIDR(input) {
		return TypeCIDR
	}

	// 3. Check for IP address
	if isIP(input) {
		return TypeIP
	}

	// 4. Check for URL (has http:// or https:// scheme)
	if isURL(input) {
		return TypeURL
	}

	// 5. Check for domain
	if isDomain(input) {
		return TypeDomain
	}

	// 6. Fallback to string
	return TypeString
}

// isCIDR checks if input is a valid CIDR notation
func isCIDR(input string) bool {
	// Quick pattern check first (faster than parsing)
	if !cidrV4Pattern.MatchString(input) && !cidrV6Pattern.MatchString(input) {
		return false
	}

	// Validate with net.ParseCIDR
	_, _, err := net.ParseCIDR(input)
	return err == nil
}

// isIP checks if input is a valid IP address (IPv4 or IPv6)
func isIP(input string) bool {
	return net.ParseIP(input) != nil
}

// isURL checks if input has http:// or https:// scheme
func isURL(input string) bool {
	lower := strings.ToLower(input)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// isDomain checks if input matches a valid domain pattern
func isDomain(input string) bool {
	return domainPattern.MatchString(input)
}

// isFile checks if the path is an existing regular file (not a directory)
func (vf *vmFunc) isFile(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	if path == "" || path == "undefined" {
		return vf.vm.ToValue(false)
	}
	info, err := os.Stat(path)
	if err != nil {
		return vf.vm.ToValue(false)
	}
	return vf.vm.ToValue(!info.IsDir())
}

// isDir checks if the path is an existing directory
func (vf *vmFunc) isDir(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	if path == "" || path == "undefined" {
		return vf.vm.ToValue(false)
	}
	info, err := os.Stat(path)
	if err != nil {
		return vf.vm.ToValue(false)
	}
	return vf.vm.ToValue(info.IsDir())
}

// isGit checks if the path is a directory containing a .git subfolder
func (vf *vmFunc) isGit(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	if path == "" || path == "undefined" {
		return vf.vm.ToValue(false)
	}
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return vf.vm.ToValue(false)
	}
	return vf.vm.ToValue(info.IsDir())
}

// isURLFunc checks if the input starts with http:// or https:// (case-insensitive)
// Named isURLFunc to avoid shadowing the existing isURL helper function.
func (vf *vmFunc) isURLFunc(call goja.FunctionCall) goja.Value {
	input := call.Argument(0).String()
	if input == "" || input == "undefined" {
		return vf.vm.ToValue(false)
	}
	return vf.vm.ToValue(isURL(input))
}

// compressedExtensions lists recognized compressed file extensions
var compressedExtensions = []string{".tar.gz", ".tar.bz2", ".tar.xz", ".tgz", ".zip", ".gz"}

// isCompress checks if the path has a compressed file extension
func (vf *vmFunc) isCompress(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	if path == "" || path == "undefined" {
		return vf.vm.ToValue(false)
	}
	lower := strings.ToLower(path)
	for _, ext := range compressedExtensions {
		if strings.HasSuffix(lower, ext) {
			return vf.vm.ToValue(true)
		}
	}
	return vf.vm.ToValue(false)
}

// Language extension mapping: file extension → language name
var langExtMap = map[string]string{
	".go":      "golang",
	".py":      "python",
	".pyw":     "python",
	".js":      "javascript",
	".mjs":     "javascript",
	".cjs":     "javascript",
	".jsx":     "javascript",
	".ts":      "typescript",
	".mts":     "typescript",
	".cts":     "typescript",
	".tsx":     "typescript",
	".rb":      "ruby",
	".java":    "java",
	".kt":      "kotlin",
	".kts":     "kotlin",
	".rs":      "rust",
	".c":       "c",
	".h":       "c",
	".cpp":     "cpp",
	".cc":      "cpp",
	".cxx":     "cpp",
	".hpp":     "cpp",
	".hxx":     "cpp",
	".cs":      "csharp",
	".swift":   "swift",
	".php":     "php",
	".pl":      "perl",
	".pm":      "perl",
	".sh":      "shell",
	".bash":    "shell",
	".zsh":     "shell",
	".lua":     "lua",
	".r":       "r",
	".scala":   "scala",
	".ex":      "elixir",
	".exs":     "elixir",
	".hs":      "haskell",
	".dart":    "dart",
	".vue":     "vue",
	".svelte":  "svelte",
}

// Shebang patterns: substring in first line → language name
var shebangMap = []struct {
	pattern  string
	language string
}{
	{"python", "python"},
	{"node", "javascript"},
	{"ruby", "ruby"},
	{"perl", "perl"},
	{"/bash", "shell"},
	{"/bin/sh", "shell"},
	{"env bash", "shell"},
	{"lua", "lua"},
}

// Directories to ignore during language detection (lowercase)
var ignoredDirs = map[string]bool{
	"test": true, "tests": true, "__tests__": true, "testdata": true,
	"test_data": true, "testing": true,
	"static": true, "public": true, "assets": true, "dist": true,
	"build": true, "out": true,
	"frontend": true, "web": true, "webapp": true, "ui": true,
	"node_modules": true, "vendor": true, "third_party": true,
	"third-party": true, "external": true,
	".git": true, ".svn": true, ".hg": true,
	"__pycache__": true, ".tox": true, ".mypy_cache": true,
}

// detectLanguage detects the dominant programming language in a folder
func (vf *vmFunc) detectLanguage(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	log := logger.Get()

	log.Debug("Calling "+terminal.HiGreen("detect_language"), zap.String("path", path))

	if path == "" || path == "undefined" {
		return vf.vm.ToValue("unknown")
	}

	result := detectFolderLanguage(path)
	log.Debug(terminal.HiGreen("detect_language")+" result", zap.String("path", path), zap.String("language", result))

	return vf.vm.ToValue(result)
}

// detectFolderLanguage walks a directory and returns the dominant programming language.
// Returns "unknown" if the path is not a directory or no source files are found.
func detectFolderLanguage(path string) string {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "unknown"
	}

	counts := make(map[string]int)

	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip entries with errors
		}

		if d.IsDir() {
			base := strings.ToLower(filepath.Base(p))
			if ignoredDirs[base] {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(p))
		if lang, ok := langExtMap[ext]; ok {
			counts[lang]++
			return nil
		}

		// No extension — check shebang
		if ext == "" {
			if lang := detectShebang(p); lang != "" {
				counts[lang]++
			}
		}

		return nil
	})

	if len(counts) == 0 {
		return "unknown"
	}

	// Find language with highest count
	bestLang := "unknown"
	bestCount := 0
	for lang, count := range counts {
		if count > bestCount {
			bestCount = count
			bestLang = lang
		}
	}

	return bestLang
}

// detectShebang reads the first line of a file and checks for a shebang pattern.
// Returns the detected language or empty string.
func detectShebang(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Limit scan to first 256 bytes
	scanner.Buffer(make([]byte, 256), 256)
	if !scanner.Scan() {
		return ""
	}
	line := scanner.Text()

	if !strings.HasPrefix(line, "#!") {
		return ""
	}

	for _, sb := range shebangMap {
		if strings.Contains(line, sb.pattern) {
			return sb.language
		}
	}

	return ""
}
