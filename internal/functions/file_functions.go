package functions

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/fileio"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// fileExists checks if a file exists
func (vf *vmFunc) fileExists(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("fileExists"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("fileExists: empty path provided")
		return vf.vm.ToValue(false)
	}

	_, err := os.Stat(path)
	exists := err == nil

	logger.Get().Debug(terminal.HiGreen("fileExists")+" result", zap.String("path", path), zap.Bool("exists", exists))
	return vf.vm.ToValue(exists)
}

// fileLength returns the number of lines in a file.
// Uses memory-mapped I/O for large files (>1MB) for better performance.
func (vf *vmFunc) fileLength(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("fileLength"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("fileLength: empty path provided")
		return vf.vm.ToValue(0)
	}

	count, err := fileio.CountNonEmptyLines(path)
	if err != nil {
		logger.Get().Warn("fileLength: failed to count lines", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(0)
	}

	logger.Get().Debug(terminal.HiGreen("fileLength")+" result", zap.String("path", path), zap.Int("count", count))
	return vf.vm.ToValue(count)
}

// dirLength returns the number of entries in a directory
func (vf *vmFunc) dirLength(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("dirLength"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("dirLength: empty path provided")
		return vf.vm.ToValue(0)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		logger.Get().Warn("dirLength: failed to read directory", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(0)
	}

	logger.Get().Debug(terminal.HiGreen("dirLength")+" result", zap.String("path", path), zap.Int("count", len(entries)))
	return vf.vm.ToValue(len(entries))
}

// fileContains checks if a file contains a pattern (regex)
func (vf *vmFunc) fileContains(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	pattern := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("fileContains"), zap.String("path", path), zap.String("pattern", pattern))

	if path == "undefined" || path == "" || pattern == "undefined" || pattern == "" {
		logger.Get().Warn("fileContains: empty path or pattern provided")
		return vf.vm.ToValue(false)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		logger.Get().Warn("fileContains: failed to read file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Try regex match first
	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("fileContains")+": invalid regex, falling back to string contains", zap.String("pattern", pattern))
		// Fall back to simple string contains
		contains := strings.Contains(string(content), pattern)
		logger.Get().Debug(terminal.HiGreen("fileContains")+" result", zap.String("path", path), zap.Bool("contains", contains))
		return vf.vm.ToValue(contains)
	}

	matches := re.MatchString(string(content))
	logger.Get().Debug(terminal.HiGreen("fileContains")+" result", zap.String("path", path), zap.Bool("matches", matches))
	return vf.vm.ToValue(matches)
}

// regexExtract extracts matching lines from a file
func (vf *vmFunc) regexExtract(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	pattern := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("regexExtract"), zap.String("path", path), zap.String("pattern", pattern))

	if path == "undefined" || path == "" || pattern == "undefined" || pattern == "" {
		logger.Get().Warn("regexExtract: empty path or pattern provided")
		return vf.vm.ToValue([]string{})
	}

	file, err := os.Open(path)
	if err != nil {
		logger.Get().Warn("regexExtract: failed to open file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue([]string{})
	}
	defer func() { _ = file.Close() }()

	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Get().Warn("regexExtract: invalid regex pattern", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue([]string{})
	}

	var matches []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, line)
		}
	}

	logger.Get().Debug(terminal.HiGreen("regexExtract")+" result", zap.String("path", path), zap.Int("matches", len(matches)))
	return vf.vm.ToValue(matches)
}

// readFile reads the entire contents of a file
func (vf *vmFunc) readFile(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("readFile"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("readFile: empty path provided")
		return vf.vm.ToValue("")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		logger.Get().Warn("readFile: failed to read file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen("readFile")+" result", zap.String("path", path), zap.Int("bytes", len(content)))
	return vf.vm.ToValue(string(content))
}

// readLines reads a file and returns an array of lines
func (vf *vmFunc) readLines(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("readLines"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("readLines: empty path provided")
		return vf.vm.ToValue([]string{})
	}

	file, err := os.Open(path)
	if err != nil {
		logger.Get().Warn("readLines: failed to open file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue([]string{})
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	logger.Get().Debug(terminal.HiGreen("readLines")+" result", zap.String("path", path), zap.Int("lines", len(lines)))
	return vf.vm.ToValue(lines)
}

func (vf *vmFunc) createFolder(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("createFolder"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("createFolder: empty path provided")
		return vf.vm.ToValue(false)
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		logger.Get().Warn("createFolder: failed to create folder", zap.String("path", path), zap.Error(err))
	}
	return vf.vm.ToValue(err == nil)
}

func (vf *vmFunc) appendFile(call goja.FunctionCall) goja.Value {
	dest := call.Argument(0).String()
	source := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("appendFile"), zap.String("dest", dest), zap.String("source", source))

	if dest == "undefined" || dest == "" || source == "undefined" || source == "" {
		logger.Get().Warn("appendFile: empty dest or source provided")
		return vf.vm.ToValue(false)
	}

	content, err := os.ReadFile(source)
	if err != nil {
		logger.Get().Warn("appendFile: failed to read source file", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		logger.Get().Warn("appendFile: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Get().Warn("appendFile: failed to open destination file", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write(content)
	if err != nil {
		logger.Get().Warn("appendFile: failed to append content", zap.String("dest", dest), zap.Error(err))
	}
	return vf.vm.ToValue(err == nil)
}

func readMatchedLines(path string, matcher func(string) bool) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if matcher(line) {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func writeLinesToFile(path string, lines []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func (vf *vmFunc) grepStringToFile(call goja.FunctionCall) goja.Value {
	dest := call.Argument(0).String()
	source := call.Argument(1).String()
	filter := call.Argument(2).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("grepStringToFile"), zap.String("dest", dest), zap.String("source", source), zap.String("filter", filter))

	if dest == "undefined" || dest == "" || source == "undefined" || source == "" || filter == "undefined" || filter == "" {
		logger.Get().Warn("grepStringToFile: empty dest, source, or filter provided")
		return vf.vm.ToValue(false)
	}

	lines, err := readMatchedLines(source, func(line string) bool {
		return strings.Contains(line, filter)
	})
	if err != nil {
		logger.Get().Warn("grepStringToFile: failed to read source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if err := writeLinesToFile(dest, lines); err != nil {
		logger.Get().Warn("grepStringToFile: failed to write destination", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}

func (vf *vmFunc) grepRegexToFile(call goja.FunctionCall) goja.Value {
	dest := call.Argument(0).String()
	source := call.Argument(1).String()
	pattern := call.Argument(2).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("grepRegexToFile"), zap.String("dest", dest), zap.String("source", source), zap.String("pattern", pattern))

	if dest == "undefined" || dest == "" || source == "undefined" || source == "" || pattern == "undefined" || pattern == "" {
		logger.Get().Warn("grepRegexToFile: empty dest, source, or pattern provided")
		return vf.vm.ToValue(false)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Get().Warn("grepRegexToFile: invalid regex pattern", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	lines, err := readMatchedLines(source, func(line string) bool {
		return re.MatchString(line)
	})
	if err != nil {
		logger.Get().Warn("grepRegexToFile: failed to read source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if err := writeLinesToFile(dest, lines); err != nil {
		logger.Get().Warn("grepRegexToFile: failed to write destination", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}

func (vf *vmFunc) grepString(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	filter := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("grepString"), zap.String("source", source), zap.String("filter", filter))

	if source == "undefined" || source == "" || filter == "undefined" || filter == "" {
		logger.Get().Warn("grepString: empty source or filter provided")
		return vf.vm.ToValue("")
	}

	lines, err := readMatchedLines(source, func(line string) bool {
		return strings.Contains(line, filter)
	})
	if err != nil {
		logger.Get().Warn("grepString: failed to read source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue("")
	}

	return vf.vm.ToValue(strings.Join(lines, "\n"))
}

func (vf *vmFunc) grepRegex(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	pattern := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("grepRegex"), zap.String("source", source), zap.String("pattern", pattern))

	if source == "undefined" || source == "" || pattern == "undefined" || pattern == "" {
		logger.Get().Warn("grepRegex: empty source or pattern provided")
		return vf.vm.ToValue("")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Get().Warn("grepRegex: invalid regex pattern", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue("")
	}

	lines, err := readMatchedLines(source, func(line string) bool {
		return re.MatchString(line)
	})
	if err != nil {
		logger.Get().Warn("grepRegex: failed to read source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue("")
	}

	return vf.vm.ToValue(strings.Join(lines, "\n"))
}

func (vf *vmFunc) glob(call goja.FunctionCall) goja.Value {
	pattern := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("glob"), zap.String("pattern", pattern))

	if pattern == "undefined" || pattern == "" {
		logger.Get().Warn("glob: empty pattern provided")
		return vf.vm.ToValue([]string{})
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		logger.Get().Warn("glob: invalid glob pattern", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue([]string{})
	}
	sort.Strings(matches)
	return vf.vm.ToValue(matches)
}

// removeBlankLines removes blank lines from a file in-place
// Usage: remove_blank_lines(path) -> bool
func (vf *vmFunc) removeBlankLines(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("removeBlankLines"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("removeBlankLines: empty path provided")
		return vf.vm.ToValue(false)
	}

	// Check if file exists and is not a directory
	info, err := os.Stat(path)
	if err != nil {
		logger.Get().Warn("removeBlankLines: file does not exist", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	if info.IsDir() {
		logger.Get().Warn("removeBlankLines: path is a directory, not a file", zap.String("path", path))
		return vf.vm.ToValue(false)
	}

	// Read the file
	file, err := os.Open(path)
	if err != nil {
		logger.Get().Warn("removeBlankLines: failed to open file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	var nonBlankLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			nonBlankLines = append(nonBlankLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		_ = file.Close()
		logger.Get().Warn("removeBlankLines: failed to read file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	_ = file.Close()

	// Write back to the same file
	content := strings.Join(nonBlankLines, "\n")
	if len(nonBlankLines) > 0 {
		content += "\n" // Add trailing newline if there are lines
	}

	if err := os.WriteFile(path, []byte(content), info.Mode()); err != nil {
		logger.Get().Warn("removeBlankLines: failed to write file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug(terminal.HiGreen("removeBlankLines")+" result", zap.String("path", path), zap.Int("lines", len(nonBlankLines)))
	return vf.vm.ToValue(true)
}

// chunkFile splits an input file into chunks of N lines each, writing chunk paths to an output manifest.
// Blank lines are skipped. Chunk files are named {base}_part_{N}{ext} in the same directory as input.
// Usage: chunk_file(input, lines_per_chunk, output) -> bool
func (vf *vmFunc) chunkFile(call goja.FunctionCall) goja.Value {
	input := call.Argument(0).String()
	chunkSize := int(call.Argument(1).ToInteger())
	output := call.Argument(2).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("chunk_file"),
		zap.String("input", input), zap.Int("chunkSize", chunkSize), zap.String("output", output))

	if input == "undefined" || input == "" || output == "undefined" || output == "" {
		logger.Get().Warn("chunk_file: input and output paths are required")
		return vf.vm.ToValue(false)
	}
	if chunkSize <= 0 {
		logger.Get().Warn("chunk_file: lines_per_chunk must be > 0", zap.Int("chunkSize", chunkSize))
		return vf.vm.ToValue(false)
	}

	// Read input file, skipping blank lines
	file, err := os.Open(input)
	if err != nil {
		logger.Get().Warn("chunk_file: failed to open input file", zap.String("input", input), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		_ = file.Close()
		logger.Get().Warn("chunk_file: failed to read input file", zap.String("input", input), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	_ = file.Close()

	if len(lines) == 0 {
		logger.Get().Debug(terminal.HiGreen("chunk_file") + ": input file has no non-blank lines")
		// Write empty manifest
		if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
			return vf.vm.ToValue(false)
		}
		if err := os.WriteFile(output, []byte(""), 0644); err != nil {
			return vf.vm.ToValue(false)
		}
		return vf.vm.ToValue(true)
	}

	// Compute chunk file naming from input path
	dir := filepath.Dir(input)
	ext := filepath.Ext(input)
	base := strings.TrimSuffix(filepath.Base(input), ext)

	var chunkPaths []string
	chunkIdx := 0

	for i := 0; i < len(lines); i += chunkSize {
		end := i + chunkSize
		if end > len(lines) {
			end = len(lines)
		}
		chunk := lines[i:end]

		chunkName := fmt.Sprintf("%s_part_%d%s", base, chunkIdx, ext)
		chunkPath := filepath.Join(dir, chunkName)

		content := strings.Join(chunk, "\n") + "\n"
		if err := os.WriteFile(chunkPath, []byte(content), 0644); err != nil {
			logger.Get().Warn("chunk_file: failed to write chunk",
				zap.String("chunkPath", chunkPath), zap.Error(err))
			return vf.vm.ToValue(false)
		}

		chunkPaths = append(chunkPaths, chunkPath)
		chunkIdx++
	}

	// Write manifest (one chunk path per line)
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		logger.Get().Warn("chunk_file: failed to create output directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	manifest := strings.Join(chunkPaths, "\n") + "\n"
	if err := os.WriteFile(output, []byte(manifest), 0644); err != nil {
		logger.Get().Warn("chunk_file: failed to write manifest", zap.String("output", output), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug(terminal.HiGreen("chunk_file")+" result",
		zap.String("input", input),
		zap.Int("totalLines", len(lines)),
		zap.Int("chunks", len(chunkPaths)),
		zap.String("output", output))
	return vf.vm.ToValue(true)
}

// zipDir creates a zip archive from a directory using Go's archive/zip
// Usage: zip_dir(source, dest) -> bool
func (vf *vmFunc) zipDir(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("zipDir"), zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("zipDir: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	// Check if source exists
	info, err := os.Stat(source)
	if err != nil {
		logger.Get().Warn("zipDir: source does not exist", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Ensure output directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Get().Warn("zipDir: failed to create output directory", zap.String("dir", dir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Create the zip file
	zipFile, err := os.Create(dest)
	if err != nil {
		logger.Get().Warn("zipDir: failed to create zip file", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = zipFile.Close() }()

	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	if info.IsDir() {
		// Walk the directory
		err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Create header
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// Calculate relative path
			relPath, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				header.Name = relPath + "/"
			} else {
				header.Name = relPath
			}

			header.Method = zip.Deflate

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() { _ = file.Close() }()

			_, err = io.Copy(writer, file)
			return err
		})
	} else {
		// Single file
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return vf.vm.ToValue(false)
		}
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return vf.vm.ToValue(false)
		}

		file, err := os.Open(source)
		if err != nil {
			return vf.vm.ToValue(false)
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(writer, file)
		if err != nil {
			logger.Get().Warn("zipDir: failed to copy file content", zap.String("source", source), zap.Error(err))
			return vf.vm.ToValue(false)
		}
	}

	success := err == nil
	if success {
		logger.Get().Debug(terminal.HiGreen("zipDir")+" result", zap.String("source", source), zap.String("dest", dest), zap.Bool("success", success))
	} else {
		logger.Get().Warn("zipDir: compression failed", zap.String("source", source), zap.Error(err))
	}
	return vf.vm.ToValue(success)
}

// unzipDir extracts a zip archive to a directory using Go's archive/zip
// Usage: unzip_dir(source, dest) -> bool
func (vf *vmFunc) unzipDir(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("unzipDir"), zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("unzipDir: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	// Open the zip file
	reader, err := zip.OpenReader(source)
	if err != nil {
		logger.Get().Warn("unzipDir: failed to open zip file", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = reader.Close() }()

	// Ensure destination exists
	if err := os.MkdirAll(dest, 0755); err != nil {
		logger.Get().Warn("unzipDir: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Extract files
	for _, file := range reader.File {
		// Sanitize the path to prevent zip slip
		destPath := filepath.Join(dest, file.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue // Skip files outside destination
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return vf.vm.ToValue(false)
			}
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return vf.vm.ToValue(false)
		}

		// Create the file
		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return vf.vm.ToValue(false)
		}

		rc, err := file.Open()
		if err != nil {
			_ = outFile.Close()
			return vf.vm.ToValue(false)
		}

		_, err = io.Copy(outFile, rc)
		_ = rc.Close()
		_ = outFile.Close()

		if err != nil {
			logger.Get().Warn("unzipDir: failed to extract file", zap.String("file", file.Name), zap.Error(err))
			return vf.vm.ToValue(false)
		}
	}

	logger.Get().Debug(terminal.HiGreen("unzipDir")+" result", zap.String("source", source), zap.String("dest", dest), zap.Bool("success", true))
	return vf.vm.ToValue(true)
}

// extractDiff compares two files and returns lines only in file2 (new content)
// Usage: extractDiff(file1, file2) -> string
func (vf *vmFunc) extractDiff(call goja.FunctionCall) goja.Value {
	file1Path := call.Argument(0).String()
	file2Path := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("extractDiff"), zap.String("file1", file1Path), zap.String("file2", file2Path))

	if file1Path == "undefined" || file1Path == "" || file2Path == "undefined" || file2Path == "" {
		logger.Get().Warn("extractDiff: empty file paths provided")
		return vf.vm.ToValue("")
	}

	// Read file1 lines into a set
	file1Lines := make(map[string]bool)
	file1, err := os.Open(file1Path)
	if err == nil {
		scanner := bufio.NewScanner(file1)
		for scanner.Scan() {
			file1Lines[scanner.Text()] = true
		}
		_ = file1.Close()
	} else {
		logger.Get().Debug(terminal.HiGreen("extractDiff")+": file1 not found, treating all lines in file2 as new", zap.String("file1", file1Path))
	}
	// If file1 doesn't exist, all lines in file2 are "new"

	// Read file2 and find lines not in file1
	file2, err := os.Open(file2Path)
	if err != nil {
		logger.Get().Warn("extractDiff: failed to open file2", zap.String("file2", file2Path), zap.Error(err))
		return vf.vm.ToValue("")
	}
	defer func() { _ = file2.Close() }()

	var newLines []string
	scanner := bufio.NewScanner(file2)
	for scanner.Scan() {
		line := scanner.Text()
		if !file1Lines[line] {
			newLines = append(newLines, line)
		}
	}

	logger.Get().Debug(terminal.HiGreen("extractDiff")+" result", zap.String("file1", file1Path), zap.String("file2", file2Path), zap.Int("new_lines", len(newLines)))
	return vf.vm.ToValue(strings.Join(newLines, "\n"))
}
