package functions

import (
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
