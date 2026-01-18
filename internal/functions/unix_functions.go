package functions

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// sortUnix sorts a file using LC_ALL=C sort -u
// Usage: sortUnix(inputFile) -> bool (sorts in-place)
// Usage: sortUnix(inputFile, outputFile) -> bool
func (vf *vmFunc) sortUnix(call goja.FunctionCall) goja.Value {
	inputFile := call.Argument(0).String()
	logger.Get().Debug("Calling sortUnix", zap.String("inputFile", inputFile))

	if inputFile == "undefined" || inputFile == "" {
		logger.Get().Warn("sortUnix: empty input file provided")
		return vf.vm.ToValue(false)
	}

	// Default: sort in-place (output = input)
	outputFile := inputFile
	if !goja.IsUndefined(call.Argument(1)) {
		out := call.Argument(1).String()
		if out != "" && out != "undefined" {
			outputFile = out
		}
	}

	// Ensure output directory exists
	if outputFile != inputFile {
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Get().Warn("sortUnix: failed to create output directory", zap.String("dir", dir), zap.Error(err))
			return vf.vm.ToValue(false)
		}
	}

	// LC_ALL=C sort -u -o output input
	// Filter out existing locale variables and set LC_ALL=C to ensure consistent sorting
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "LC_ALL=") && !strings.HasPrefix(e, "LC_COLLATE=") && !strings.HasPrefix(e, "LANG=") {
			env = append(env, e)
		}
	}
	env = append(env, "LC_ALL=C")

	cmd := exec.Command("sort", "-u", "-o", outputFile, inputFile)
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("sortUnix: sort command failed", zap.String("inputFile", inputFile), zap.Error(err))
	} else {
		logger.Get().Debug("sortUnix result", zap.String("inputFile", inputFile), zap.String("outputFile", outputFile), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// wgetUnix downloads a file using wget
// Usage: wgetUnix(url) -> bool (saves to current directory)
// Usage: wgetUnix(url, outputPath) -> bool
func (vf *vmFunc) wgetUnix(call goja.FunctionCall) goja.Value {
	url := call.Argument(0).String()
	outputPath := call.Argument(1).String()
	logger.Get().Debug("Calling wgetUnix", zap.String("url", url), zap.String("outputPath", outputPath))

	if url == "undefined" || url == "" {
		logger.Get().Warn("wgetUnix: empty URL provided")
		return vf.vm.ToValue(false)
	}

	args := []string{"-q", "--no-check-certificate"}

	if outputPath != "undefined" && outputPath != "" {
		// Ensure directory exists
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Get().Warn("wgetUnix: failed to create output directory", zap.String("dir", dir), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		args = append(args, "-O", outputPath)
	}
	args = append(args, url)

	cmd := exec.Command("wget", args...)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("wgetUnix: wget command failed", zap.String("url", url), zap.Error(err))
	} else {
		logger.Get().Debug("wgetUnix result", zap.String("url", url), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// injectGitHubTokenForClone injects GitHub token into GitHub URLs for authentication
// Priority: GITHUB_API_KEY (from settings) > GH_TOKEN (from OS env)
func injectGitHubTokenForClone(repoURL string) string {
	// First: try GITHUB_API_KEY from settings (exported to env by root.go)
	token := os.Getenv("GITHUB_API_KEY")
	if token == "" {
		// Fallback: GH_TOKEN from OS environment (used by GitHub CLI)
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		return repoURL
	}
	// Only inject for github.com HTTPS URLs
	if strings.Contains(repoURL, "github.com") && strings.HasPrefix(repoURL, "https://") {
		repoURL = strings.Replace(repoURL, "https://github.com", "https://"+token+"@github.com", 1)
	}
	return repoURL
}

// gitClone clones a git repository
// Usage: gitClone(repo) -> bool (clones to current directory)
// Usage: gitClone(repo, dest) -> bool
func (vf *vmFunc) gitClone(call goja.FunctionCall) goja.Value {
	repo := call.Argument(0).String()
	// Inject GitHub token for private repo access
	repo = injectGitHubTokenForClone(repo)
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling gitClone", zap.String("repo", repo), zap.String("dest", dest))

	if repo == "undefined" || repo == "" {
		logger.Get().Warn("gitClone: empty repo provided")
		return vf.vm.ToValue(false)
	}

	args := []string{"clone", "--depth", "1"}
	args = append(args, repo)

	if dest != "undefined" && dest != "" {
		// Ensure parent directory exists
		dir := filepath.Dir(dest)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Get().Warn("gitClone: failed to create parent directory", zap.String("dir", dir), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		args = append(args, dest)
	}

	cmd := exec.Command("git", args...)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("gitClone: git clone failed", zap.String("repo", repo), zap.Error(err))
	} else {
		logger.Get().Debug("gitClone result", zap.String("repo", repo), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// zipUnix creates a zip archive using the zip command
// Usage: zipUnix(source, dest) -> bool (zip -r dest source)
func (vf *vmFunc) zipUnix(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling zipUnix", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("zipUnix: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	// Ensure output directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Get().Warn("zipUnix: failed to create output directory", zap.String("dir", dir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// zip -r dest source
	cmd := exec.Command("zip", "-r", dest, source)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("zipUnix: zip command failed", zap.String("source", source), zap.Error(err))
	} else {
		logger.Get().Debug("zipUnix result", zap.String("source", source), zap.String("dest", dest), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// unzipUnix extracts a zip archive using the unzip command
// Usage: unzipUnix(source) -> bool (extracts to current directory)
// Usage: unzipUnix(source, dest) -> bool (unzip source -d dest)
func (vf *vmFunc) unzipUnix(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling unzipUnix", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" {
		logger.Get().Warn("unzipUnix: empty source provided")
		return vf.vm.ToValue(false)
	}

	args := []string{"-o", source} // -o for overwrite without prompting

	if dest != "undefined" && dest != "" {
		// Ensure destination directory exists
		if err := os.MkdirAll(dest, 0755); err != nil {
			logger.Get().Warn("unzipUnix: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		args = append(args, "-d", dest)
	}

	cmd := exec.Command("unzip", args...)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("unzipUnix: unzip command failed", zap.String("source", source), zap.Error(err))
	} else {
		logger.Get().Debug("unzipUnix result", zap.String("source", source), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// tarUnix creates a tar.gz archive using the tar command
// Usage: tarUnix(source, dest) -> bool (tar -czf dest source)
func (vf *vmFunc) tarUnix(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling tarUnix", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("tarUnix: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	// Ensure output directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Get().Warn("tarUnix: failed to create output directory", zap.String("dir", dir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// tar -czf dest source
	cmd := exec.Command("tar", "-czf", dest, source)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("tarUnix: tar command failed", zap.String("source", source), zap.Error(err))
	} else {
		logger.Get().Debug("tarUnix result", zap.String("source", source), zap.String("dest", dest), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// untarUnix extracts a tar.gz archive using the tar command
// Usage: untarUnix(source) -> bool (extracts to current directory)
// Usage: untarUnix(source, dest) -> bool (tar -xzf source -C dest)
func (vf *vmFunc) untarUnix(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling untarUnix", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" {
		logger.Get().Warn("untarUnix: empty source provided")
		return vf.vm.ToValue(false)
	}

	args := []string{"-xzf", source}

	if dest != "undefined" && dest != "" {
		// Ensure destination directory exists
		if err := os.MkdirAll(dest, 0755); err != nil {
			logger.Get().Warn("untarUnix: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		args = append(args, "-C", dest)
	}

	cmd := exec.Command("tar", args...)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("untarUnix: tar extract failed", zap.String("source", source), zap.Error(err))
	} else {
		logger.Get().Debug("untarUnix result", zap.String("source", source), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// diffUnix compares two files using the diff command
// Usage: diffUnix(file1, file2) -> structured JSON response
// Usage: diffUnix(file1, file2, output) -> structured JSON response (also saves to output file)
// Returns: {error, message, output, file1, file2, line_count}
func (vf *vmFunc) diffUnix(call goja.FunctionCall) goja.Value {
	file1 := call.Argument(0).String()
	file2 := call.Argument(1).String()
	logger.Get().Debug("Calling diffUnix", zap.String("file1", file1), zap.String("file2", file2))

	if file1 == "undefined" || file1 == "" || file2 == "undefined" || file2 == "" {
		logger.Get().Warn("diffUnix: empty file paths provided")
		return vf.vm.ToValue(map[string]interface{}{
			"error":      "empty file paths provided",
			"message":    "error",
			"output":     "",
			"file1":      file1,
			"file2":      file2,
			"line_count": 0,
		})
	}

	// diff file1 file2 (returns non-zero exit code if files differ, which is expected)
	cmd := exec.Command("diff", file1, file2)
	output, _ := cmd.Output() // Ignore error since diff returns 1 when files differ

	outputStr := string(output)

	// Count lines in diff output
	lineCount := 0
	if len(outputStr) > 0 {
		lineCount = strings.Count(outputStr, "\n")
		if !strings.HasSuffix(outputStr, "\n") {
			lineCount++
		}
	}

	// If output file specified, save the diff
	outputPath := call.Argument(2).String()
	if outputPath != "undefined" && outputPath != "" {
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err == nil {
			_ = os.WriteFile(outputPath, output, 0644)
		}
	}

	logger.Get().Debug("diffUnix result", zap.String("file1", file1), zap.String("file2", file2), zap.Int("line_count", lineCount))
	return vf.vm.ToValue(map[string]interface{}{
		"error":      nil,
		"message":    "success",
		"output":     outputStr,
		"file1":      file1,
		"file2":      file2,
		"line_count": lineCount,
	})
}

// parseSedSyntax parses sed s/pattern/replacement/flags syntax
// Returns pattern, replacement, global flag, and error
func parseSedSyntax(syntax string) (pattern, replacement string, global bool, err error) {
	syntax = strings.TrimSpace(syntax)

	// Must start with 's'
	if len(syntax) < 4 || syntax[0] != 's' {
		return "", "", false, fmt.Errorf("invalid sed syntax: must start with 's'")
	}

	// Get the delimiter (character after 's')
	delim := syntax[1]

	// Find the parts: s/pattern/replacement/flags
	// Start after "s<delim>"
	rest := syntax[2:]

	// Find pattern end (first unescaped delimiter)
	patternEnd := findUnescapedDelim(rest, delim)
	if patternEnd == -1 {
		return "", "", false, fmt.Errorf("invalid sed syntax: missing pattern delimiter")
	}

	pattern = rest[:patternEnd]
	rest = rest[patternEnd+1:]

	// Find replacement end (next unescaped delimiter)
	replacementEnd := findUnescapedDelim(rest, delim)
	if replacementEnd == -1 {
		// No closing delimiter, rest is the replacement with no flags
		replacement = rest
		return pattern, replacement, false, nil
	}

	replacement = rest[:replacementEnd]
	flags := rest[replacementEnd+1:]

	// Check for global flag
	global = strings.Contains(flags, "g")

	return pattern, replacement, global, nil
}

// findUnescapedDelim finds the first unescaped occurrence of delimiter
func findUnescapedDelim(s string, delim byte) int {
	escaped := false
	for i := 0; i < len(s); i++ {
		if escaped {
			escaped = false
			continue
		}
		if s[i] == '\\' {
			escaped = true
			continue
		}
		if s[i] == delim {
			return i
		}
	}
	return -1
}

// sedStringReplace performs literal string replacement using sed-like syntax
// Usage: sed_string_replace('s/old/new/g', '/path/to/source', '/path/to/dest') -> bool
func (vf *vmFunc) sedStringReplace(call goja.FunctionCall) goja.Value {
	sedSyntax := call.Argument(0).String()
	sourceFile := call.Argument(1).String()
	destFile := call.Argument(2).String()

	logger.Get().Debug("Calling sedStringReplace",
		zap.String("syntax", sedSyntax),
		zap.String("source", sourceFile),
		zap.String("dest", destFile))

	if sedSyntax == "undefined" || sedSyntax == "" {
		logger.Get().Warn("sedStringReplace: empty sed syntax provided")
		return vf.vm.ToValue(false)
	}

	if sourceFile == "undefined" || sourceFile == "" {
		logger.Get().Warn("sedStringReplace: empty source file provided")
		return vf.vm.ToValue(false)
	}

	if destFile == "undefined" || destFile == "" {
		logger.Get().Warn("sedStringReplace: empty dest file provided")
		return vf.vm.ToValue(false)
	}

	// Parse sed syntax
	pattern, replacement, global, err := parseSedSyntax(sedSyntax)
	if err != nil {
		logger.Get().Warn("sedStringReplace: failed to parse sed syntax", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Open source file
	source, err := os.Open(sourceFile)
	if err != nil {
		logger.Get().Warn("sedStringReplace: failed to open source file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = source.Close() }()

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		logger.Get().Warn("sedStringReplace: failed to create dest directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Create dest file
	dest, err := os.Create(destFile)
	if err != nil {
		logger.Get().Warn("sedStringReplace: failed to create dest file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = dest.Close() }()

	// Stream process line by line
	scanner := bufio.NewScanner(source)
	writer := bufio.NewWriter(dest)

	for scanner.Scan() {
		line := scanner.Text()
		var newLine string
		if global {
			newLine = strings.ReplaceAll(line, pattern, replacement)
		} else {
			newLine = strings.Replace(line, pattern, replacement, 1)
		}
		_, _ = writer.WriteString(newLine + "\n")
	}

	if err := scanner.Err(); err != nil {
		logger.Get().Warn("sedStringReplace: error reading source file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if err := writer.Flush(); err != nil {
		logger.Get().Warn("sedStringReplace: error writing dest file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("sedStringReplace completed successfully",
		zap.String("source", sourceFile),
		zap.String("dest", destFile))

	return vf.vm.ToValue(true)
}

// sedRegexReplace performs regex replacement using sed-like syntax
// Usage: sed_regex_replace('s/[0-9]+/NUM/g', '/path/to/source', '/path/to/dest') -> bool
func (vf *vmFunc) sedRegexReplace(call goja.FunctionCall) goja.Value {
	sedSyntax := call.Argument(0).String()
	sourceFile := call.Argument(1).String()
	destFile := call.Argument(2).String()

	logger.Get().Debug("Calling sedRegexReplace",
		zap.String("syntax", sedSyntax),
		zap.String("source", sourceFile),
		zap.String("dest", destFile))

	if sedSyntax == "undefined" || sedSyntax == "" {
		logger.Get().Warn("sedRegexReplace: empty sed syntax provided")
		return vf.vm.ToValue(false)
	}

	if sourceFile == "undefined" || sourceFile == "" {
		logger.Get().Warn("sedRegexReplace: empty source file provided")
		return vf.vm.ToValue(false)
	}

	if destFile == "undefined" || destFile == "" {
		logger.Get().Warn("sedRegexReplace: empty dest file provided")
		return vf.vm.ToValue(false)
	}

	// Parse sed syntax
	pattern, replacement, global, err := parseSedSyntax(sedSyntax)
	if err != nil {
		logger.Get().Warn("sedRegexReplace: failed to parse sed syntax", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Compile regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Get().Warn("sedRegexReplace: failed to compile regex", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Open source file
	source, err := os.Open(sourceFile)
	if err != nil {
		logger.Get().Warn("sedRegexReplace: failed to open source file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = source.Close() }()

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		logger.Get().Warn("sedRegexReplace: failed to create dest directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Create dest file
	dest, err := os.Create(destFile)
	if err != nil {
		logger.Get().Warn("sedRegexReplace: failed to create dest file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = dest.Close() }()

	// Stream process line by line
	scanner := bufio.NewScanner(source)
	writer := bufio.NewWriter(dest)

	for scanner.Scan() {
		line := scanner.Text()
		var newLine string
		if global {
			newLine = re.ReplaceAllString(line, replacement)
		} else {
			// Replace only first occurrence
			loc := re.FindStringIndex(line)
			if loc != nil {
				newLine = line[:loc[0]] + re.ReplaceAllString(line[loc[0]:loc[1]], replacement) + line[loc[1]:]
			} else {
				newLine = line
			}
		}
		_, _ = writer.WriteString(newLine + "\n")
	}

	if err := scanner.Err(); err != nil {
		logger.Get().Warn("sedRegexReplace: error reading source file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if err := writer.Flush(); err != nil {
		logger.Get().Warn("sedRegexReplace: error writing dest file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("sedRegexReplace completed successfully",
		zap.String("source", sourceFile),
		zap.String("dest", destFile))

	return vf.vm.ToValue(true)
}
