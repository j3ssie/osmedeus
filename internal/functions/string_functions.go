package functions

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// Subdomain cleaning regex patterns (compiled once for performance)
var (
	cleanSubDomainRE   = regexp.MustCompile(`(([a-zA-Z0-9]{1}|[_a-zA-Z0-9]{1}[_a-zA-Z0-9-]{0,61}[a-zA-Z0-9]{1})[.]{1})+[a-zA-Z]{2,61}`)
	cleanSubStripRE    = regexp.MustCompile(`^(?:u[0-9a-f]{4}|20|22|25|2b|2f|3d|3a|40)`)
	cleanSubIPv4RE     = regexp.MustCompile(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.`)
	cleanSubIPv4DashRE = regexp.MustCompile(`[0-9]{1,3}-[0-9]{1,3}-[0-9]{1,3}`)
)

// trim trims whitespace from a string
func (vf *vmFunc) trim(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("trim"), zap.Int("inputLength", len(s)))

	if s == "undefined" {
		return vf.vm.ToValue("")
	}

	return vf.vm.ToValue(strings.TrimSpace(s))
}

// split splits a string by delimiter
func (vf *vmFunc) split(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	delimiter := call.Argument(1).String()

	if s == "undefined" || delimiter == "undefined" {
		return vf.vm.ToValue([]string{})
	}

	parts := strings.Split(s, delimiter)
	return vf.vm.ToValue(parts)
}

// join joins an array with a delimiter
func (vf *vmFunc) join(call goja.FunctionCall) goja.Value {
	arrValue := call.Argument(0)
	delimiter := call.Argument(1).String()

	if delimiter == "undefined" {
		delimiter = ""
	}

	exported := arrValue.Export()
	if exported == nil {
		return vf.vm.ToValue("")
	}

	// Handle different array types
	var parts []string
	switch v := exported.(type) {
	case []string:
		parts = v
	case []interface{}:
		for _, item := range v {
			parts = append(parts, toString(item))
		}
	default:
		return vf.vm.ToValue("")
	}

	return vf.vm.ToValue(strings.Join(parts, delimiter))
}

// replace replaces occurrences in a string
func (vf *vmFunc) replace(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	old := call.Argument(1).String()
	new := call.Argument(2).String()

	if s == "undefined" || old == "undefined" {
		return vf.vm.ToValue(s)
	}
	if new == "undefined" {
		new = ""
	}

	return vf.vm.ToValue(strings.ReplaceAll(s, old, new))
}

// contains checks if a string contains a substring
func (vf *vmFunc) contains(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	substr := call.Argument(1).String()

	if s == "undefined" || substr == "undefined" {
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(strings.Contains(s, substr))
}

// startsWith checks if a string starts with a prefix
func (vf *vmFunc) startsWith(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	prefix := call.Argument(1).String()

	if s == "undefined" || prefix == "undefined" {
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(strings.HasPrefix(s, prefix))
}

// endsWith checks if a string ends with a suffix
func (vf *vmFunc) endsWith(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	suffix := call.Argument(1).String()

	if s == "undefined" || suffix == "undefined" {
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(strings.HasSuffix(s, suffix))
}

// toLowerCase converts a string to lowercase
func (vf *vmFunc) toLowerCase(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	if s == "undefined" {
		return vf.vm.ToValue("")
	}

	return vf.vm.ToValue(strings.ToLower(s))
}

// toUpperCase converts a string to uppercase
func (vf *vmFunc) toUpperCase(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	if s == "undefined" {
		return vf.vm.ToValue("")
	}

	return vf.vm.ToValue(strings.ToUpper(s))
}

// match checks if a string matches a regex pattern
func (vf *vmFunc) match(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	pattern := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("match"), zap.String("pattern", pattern), zap.Int("inputLength", len(s)))

	if s == "undefined" || pattern == "undefined" {
		logger.Get().Warn("match: undefined input or pattern")
		return vf.vm.ToValue(false)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Get().Warn("match: invalid regex pattern", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	matches := re.MatchString(s)
	logger.Get().Debug(terminal.HiGreen("match")+" result", zap.String("pattern", pattern), zap.Bool("matches", matches))
	return vf.vm.ToValue(matches)
}

// regexMatch checks if a string matches a regex pattern (pattern first)
// Usage: regex_match(pattern, string) -> bool
func (vf *vmFunc) regexMatch(call goja.FunctionCall) goja.Value {
	pattern := call.Argument(0).String()
	s := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("regexMatch"), zap.String("pattern", pattern), zap.Int("inputLength", len(s)))

	if pattern == "undefined" || s == "undefined" {
		logger.Get().Warn("regexMatch: undefined pattern or input")
		return vf.vm.ToValue(false)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Get().Warn("regexMatch: invalid regex pattern", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	matches := re.MatchString(s)
	logger.Get().Debug(terminal.HiGreen("regexMatch")+" result", zap.String("pattern", pattern), zap.Bool("matches", matches))
	return vf.vm.ToValue(matches)
}

// cutWithDelim extracts a field from input based on delimiter (1-indexed like cut)
// Usage: cut_with_delim(input, delim, field) -> string
func (vf *vmFunc) cutWithDelim(call goja.FunctionCall) goja.Value {
	input := call.Argument(0).String()
	delim := call.Argument(1).String()
	field := call.Argument(2).ToInteger()

	if input == "undefined" || delim == "undefined" {
		return vf.vm.ToValue("")
	}

	parts := strings.Split(input, delim)

	// Field is 1-indexed (like cut command)
	idx := int(field) - 1
	if idx < 0 || idx >= len(parts) {
		return vf.vm.ToValue("")
	}

	return vf.vm.ToValue(parts[idx])
}

// normalizePath replaces special characters with underscore for clean directory/file names
// Replaces: / | : \ * ? " < > with _
// Usage: normalize_path(input) -> string
func (vf *vmFunc) normalizePath(call goja.FunctionCall) goja.Value {
	input := call.Argument(0).String()

	if input == "undefined" || input == "" {
		return vf.vm.ToValue("")
	}

	// Characters to replace with underscore
	replacer := strings.NewReplacer(
		"/", "_",
		"|", "_",
		":", "_",
		"\\", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
	)

	normalized := replacer.Replace(input)

	return vf.vm.ToValue(normalized)
}

// cleanSub cleans and deduplicates subdomains in a file (in-place)
// Extracts valid subdomains, removes duplicates, filters IP-like patterns
// Optional target parameter filters to only include subdomains of that domain
// Optimized for huge files using streaming I/O
// Usage: clean_sub(path, target?) -> bool
func (vf *vmFunc) cleanSub(call goja.FunctionCall) goja.Value {
	filePath := call.Argument(0).String()
	target := call.Argument(1).String()
	log := logger.Get()

	// Normalize target domain
	if target == "undefined" {
		target = ""
	} else {
		target = strings.ToLower(strings.TrimSpace(target))
	}

	log.Debug("Calling "+terminal.HiGreen("clean_sub"), zap.String("path", filePath), zap.String("target", target))

	if filePath == "undefined" || filePath == "" {
		log.Warn("clean_sub: empty path provided")
		return vf.vm.ToValue(false)
	}

	// Open source file
	f, err := os.Open(filePath)
	if err != nil {
		log.Warn("clean_sub: failed to open file", zap.String("path", filePath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Create temp file in same directory for atomic replacement
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), ".clean_sub_*.tmp")
	if err != nil {
		_ = f.Close()
		log.Warn("clean_sub: failed to create temp file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	tempPath := tempFile.Name()

	// Use buffered I/O for performance on large files
	writer := bufio.NewWriterSize(tempFile, 256*1024)
	seen := make(map[string]struct{}, 10000)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		cleaned := cleanSubdomainLine(line)
		if cleaned == "" {
			continue
		}

		// Filter by target domain if specified
		if target != "" {
			// Accept: target.com or *.target.com (subdomain of target)
			if cleaned != target && !strings.HasSuffix(cleaned, "."+target) {
				continue
			}
		}

		if _, exists := seen[cleaned]; exists {
			continue
		}
		seen[cleaned] = struct{}{}
		_, _ = writer.WriteString(cleaned)
		_ = writer.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		_ = tempFile.Close()
		_ = f.Close()
		_ = os.Remove(tempPath)
		log.Warn("clean_sub: scan error", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	_ = writer.Flush()
	_ = tempFile.Close()
	_ = f.Close()

	// Atomic replace: rename temp file to original
	if err := os.Rename(tempPath, filePath); err != nil {
		_ = os.Remove(tempPath)
		log.Warn("clean_sub: failed to replace file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	log.Debug(terminal.HiGreen("clean_sub")+": completed", zap.Int("unique_count", len(seen)))
	return vf.vm.ToValue(true)
}

// cleanSubdomainLine extracts and cleans a subdomain from a line
func cleanSubdomainLine(line string) string {
	// Extract subdomain using regex
	name := cleanSubDomainRE.FindString(line)
	if name == "" {
		return ""
	}
	name = strings.ToLower(name)

	// Strip encoded/weird chars (unicode escapes, URL-encoded)
	for {
		name = strings.Trim(name, "-.")
		if idx := cleanSubStripRE.FindStringIndex(name); idx != nil {
			name = name[idx[1]:]
		} else {
			break
		}
	}

	// Remove asterisk wildcard label (*.domain.com -> domain.com)
	if idx := strings.LastIndex(name, "*."); idx != -1 {
		name = name[idx+2:]
	}

	// Filter out IP-like patterns (e.g., 192.168.1.example.com or 1-2-3.example.com)
	if cleanSubIPv4RE.MatchString(name) || cleanSubIPv4DashRE.MatchString(name) {
		return ""
	}

	return name
}

// parseInt parses a string to integer
func (vf *vmFunc) parseInt(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("parseInt"), zap.String("input", s))

	if s == "undefined" {
		logger.Get().Warn("parseInt: undefined input")
		return vf.vm.ToValue(0)
	}

	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		logger.Get().Warn("parseInt: failed to parse", zap.String("input", s), zap.Error(err))
		return vf.vm.ToValue(0)
	}

	logger.Get().Debug(terminal.HiGreen("parseInt")+" result", zap.Int("value", i))
	return vf.vm.ToValue(i)
}

// parseFloat parses a string to float
func (vf *vmFunc) parseFloat(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("parseFloat"), zap.String("input", s))

	if s == "undefined" {
		logger.Get().Warn("parseFloat: undefined input")
		return vf.vm.ToValue(0.0)
	}

	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		logger.Get().Warn("parseFloat: failed to parse", zap.String("input", s), zap.Error(err))
		return vf.vm.ToValue(0.0)
	}

	logger.Get().Debug(terminal.HiGreen("parseFloat")+" result", zap.Float64("value", f))
	return vf.vm.ToValue(f)
}

// toStringJS converts a value to string (renamed to avoid conflict)
func (vf *vmFunc) toString(call goja.FunctionCall) goja.Value {
	v := call.Argument(0)
	return vf.vm.ToValue(v.String())
}

// toBoolean converts a value to boolean
func (vf *vmFunc) toBoolean(call goja.FunctionCall) goja.Value {
	v := call.Argument(0)
	b := v.ToBoolean()
	return vf.vm.ToValue(b)
}

// length returns the length of a string or array
func (vf *vmFunc) length(call goja.FunctionCall) goja.Value {
	v := call.Argument(0)

	// Try as string first
	if goja.IsString(v) {
		s := v.String()
		return vf.vm.ToValue(len(s))
	}

	// Try as array
	exported := v.Export()
	if exported == nil {
		return vf.vm.ToValue(0)
	}

	switch arr := exported.(type) {
	case []string:
		return vf.vm.ToValue(len(arr))
	case []interface{}:
		return vf.vm.ToValue(len(arr))
	default:
		return vf.vm.ToValue(0)
	}
}

// isEmpty checks if a value is empty
func (vf *vmFunc) isEmpty(call goja.FunctionCall) goja.Value {
	v := call.Argument(0)

	if goja.IsUndefined(v) || goja.IsNull(v) {
		return vf.vm.ToValue(true)
	}

	if goja.IsString(v) {
		s := strings.TrimSpace(v.String())
		return vf.vm.ToValue(s == "" || s == "undefined")
	}

	exported := v.Export()
	if exported == nil {
		return vf.vm.ToValue(true)
	}

	switch arr := exported.(type) {
	case []string:
		return vf.vm.ToValue(len(arr) == 0)
	case []interface{}:
		return vf.vm.ToValue(len(arr) == 0)
	default:
		return vf.vm.ToValue(false)
	}
}

// isNotEmpty checks if a value is not empty
func (vf *vmFunc) isNotEmpty(call goja.FunctionCall) goja.Value {
	isEmpty := vf.isEmpty(call)
	b := isEmpty.ToBoolean()
	return vf.vm.ToValue(!b)
}

// Helper function to convert interface to string
func toString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(s)
	default:
		return ""
	}
}
