package functions

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
)

// renderMarkdownFromFile reads a markdown file and renders it with terminal styling
// Usage: render_markdown_from_file(path) -> string
func (vf *vmFunc) renderMarkdownFromFile(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling renderMarkdownFromFile", zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("renderMarkdownFromFile: empty path provided")
		return vf.vm.ToValue("")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		logger.Get().Warn("renderMarkdownFromFile: failed to read file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue("")
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
	)
	if err != nil {
		logger.Get().Warn("renderMarkdownFromFile: failed to create renderer", zap.Error(err))
		return vf.vm.ToValue(string(content))
	}

	rendered, err := renderer.Render(string(content))
	if err != nil {
		logger.Get().Warn("renderMarkdownFromFile: failed to render markdown", zap.Error(err))
		return vf.vm.ToValue(string(content))
	}

	logger.Get().Debug("renderMarkdownFromFile result", zap.String("path", path), zap.Int("renderedLength", len(rendered)))
	return vf.vm.ToValue(rendered)
}

// printMarkdownFromFile reads a markdown file, prints it with syntax highlighting, and returns the rendered content
// Usage: print_markdown_from_file(path) -> string
func (vf *vmFunc) printMarkdownFromFile(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling printMarkdownFromFile", zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("printMarkdownFromFile: empty path provided")
		return vf.vm.ToValue("")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		logger.Get().Warn("printMarkdownFromFile: failed to read file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue("")
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
	)
	if err != nil {
		logger.Get().Warn("printMarkdownFromFile: failed to create renderer", zap.Error(err))
		fmt.Print(string(content))
		return vf.vm.ToValue(string(content))
	}

	rendered, err := renderer.Render(string(content))
	if err != nil {
		logger.Get().Warn("printMarkdownFromFile: failed to render markdown", zap.Error(err))
		fmt.Print(string(content))
		return vf.vm.ToValue(string(content))
	}

	logger.Get().Debug("printMarkdownFromFile result", zap.String("path", path), zap.Int("renderedLength", len(rendered)))
	fmt.Print(rendered)
	return vf.vm.ToValue(rendered)
}

// convertJSONLToMarkdown reads a JSONL file and converts it to a markdown table
// Usage: convert_jsonl_to_markdown(input_path, output_path) -> bool
func (vf *vmFunc) convertJSONLToMarkdown(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		logger.Get().Warn("convertJSONLToMarkdown: requires 2 arguments (input_path, output_path)")
		return vf.vm.ToValue(false)
	}

	inputPath := call.Argument(0).String()
	outputPath := call.Argument(1).String()
	logger.Get().Debug("Calling convertJSONLToMarkdown", zap.String("input", inputPath), zap.String("output", outputPath))

	if inputPath == "undefined" || inputPath == "" {
		logger.Get().Warn("convertJSONLToMarkdown: empty input path provided")
		return vf.vm.ToValue(false)
	}

	if outputPath == "undefined" || outputPath == "" {
		logger.Get().Warn("convertJSONLToMarkdown: empty output path provided")
		return vf.vm.ToValue(false)
	}

	file, err := os.Open(inputPath)
	if err != nil {
		logger.Get().Warn("convertJSONLToMarkdown: failed to open file", zap.String("path", inputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = file.Close() }()

	// Collect headers preserving order from first object, then append any new keys
	var headers []string
	keySet := make(map[string]bool)
	var p fastjson.Parser

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		v, err := p.Parse(line)
		if err != nil {
			logger.Get().Debug("convertJSONLToMarkdown: skipping invalid JSON line", zap.Error(err))
			continue
		}
		o, err := v.Object()
		if err != nil || o == nil {
			continue
		}
		// Visit preserves order within each JSON object
		o.Visit(func(k []byte, _ *fastjson.Value) {
			key := string(k)
			if !keySet[key] {
				keySet[key] = true
				headers = append(headers, key)
			}
		})
	}
	if err := scanner.Err(); err != nil {
		logger.Get().Warn("convertJSONLToMarkdown: failed to read file", zap.String("path", inputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if len(headers) == 0 {
		logger.Get().Debug("convertJSONLToMarkdown: no valid rows found", zap.String("path", inputPath))
		return vf.vm.ToValue(false)
	}

	// Build markdown table
	var sb strings.Builder

	// Header row
	sb.WriteString("| ")
	sb.WriteString(strings.Join(headers, " | "))
	sb.WriteString(" |\n")

	// Separator row
	sb.WriteString("|")
	for range headers {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	if _, err := file.Seek(0, 0); err != nil {
		logger.Get().Warn("convertJSONLToMarkdown: failed to rewind file", zap.String("path", inputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	rowCount := 0
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		v, err := p.Parse(line)
		if err != nil {
			continue
		}
		obj, err := v.Object()
		if err != nil || obj == nil {
			continue
		}
		rowCount++

		sb.WriteString("| ")
		for i, header := range headers {
			if i > 0 {
				sb.WriteString(" | ")
			}
			val := obj.Get(header)
			if val != nil {
				sb.WriteString(formatValueFromFastjson(val))
			}
		}
		sb.WriteString(" |\n")
	}
	if err := scanner.Err(); err != nil {
		logger.Get().Warn("convertJSONLToMarkdown: failed to read file", zap.String("path", inputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Get().Warn("convertJSONLToMarkdown: failed to create output directory",
			zap.String("path", outputDir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		logger.Get().Warn("convertJSONLToMarkdown: failed to write output",
			zap.String("path", outputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("convertJSONLToMarkdown result", zap.String("input", inputPath), zap.String("output", outputPath), zap.Int("rows", rowCount), zap.Int("columns", len(headers)))
	return vf.vm.ToValue(true)
}

// convertCSVToMarkdown reads a CSV file and converts it to a markdown table
// Usage: convert_csv_to_markdown(path) -> string
func (vf *vmFunc) convertCSVToMarkdown(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling convertCSVToMarkdown", zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("convertCSVToMarkdown: empty path provided")
		return vf.vm.ToValue("")
	}

	file, err := os.Open(path)
	if err != nil {
		logger.Get().Warn("convertCSVToMarkdown: failed to open file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue("")
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		logger.Get().Warn("convertCSVToMarkdown: failed to parse CSV", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue("")
	}

	if len(records) == 0 {
		logger.Get().Debug("convertCSVToMarkdown: no records found", zap.String("path", path))
		return vf.vm.ToValue("")
	}

	var sb strings.Builder

	// Header row (first line)
	headers := records[0]
	sb.WriteString("| ")
	sb.WriteString(strings.Join(headers, " | "))
	sb.WriteString(" |\n")

	// Separator row
	sb.WriteString("|")
	for range headers {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range records[1:] {
		sb.WriteString("| ")
		for i, cell := range row {
			if i > 0 {
				sb.WriteString(" | ")
			}
			// Escape pipe characters in cell content
			sb.WriteString(strings.ReplaceAll(cell, "|", "\\|"))
		}
		sb.WriteString(" |\n")
	}

	logger.Get().Debug("convertCSVToMarkdown result", zap.String("path", path), zap.Int("rows", len(records)-1), zap.Int("columns", len(headers)))
	return vf.vm.ToValue(sb.String())
}

func formatValueForCSVFromFastjson(val *fastjson.Value) string {
	if val == nil {
		return ""
	}
	switch val.Type() {
	case fastjson.TypeString:
		return string(val.GetStringBytes())
	case fastjson.TypeNull:
		return ""
	default:
		return string(val.MarshalTo(nil))
	}
}

func (vf *vmFunc) jsonlToCSV(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling jsonlToCSV", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("jsonlToCSV: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	f, err := os.Open(source)
	if err != nil {
		logger.Get().Warn("jsonlToCSV: failed to open source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = f.Close() }()

	// Collect headers preserving order from first object, then append any new keys
	var headers []string
	keySet := make(map[string]bool)
	var p fastjson.Parser

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		v, err := p.Parse(line)
		if err != nil {
			logger.Get().Debug("jsonlToCSV: skipping invalid JSON line", zap.Error(err))
			continue
		}
		o, err := v.Object()
		if err != nil || o == nil {
			continue
		}
		// Visit preserves order within each JSON object
		o.Visit(func(k []byte, _ *fastjson.Value) {
			key := string(k)
			if !keySet[key] {
				keySet[key] = true
				headers = append(headers, key)
			}
		})
	}
	if err := scanner.Err(); err != nil {
		logger.Get().Warn("jsonlToCSV: failed to read source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if len(headers) == 0 {
		return vf.vm.ToValue(false)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		logger.Get().Warn("jsonlToCSV: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	out, err := os.Create(dest)
	if err != nil {
		logger.Get().Warn("jsonlToCSV: failed to create destination file", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = out.Close() }()

	w := csv.NewWriter(out)
	if err := w.Write(headers); err != nil {
		logger.Get().Warn("jsonlToCSV: failed to write CSV headers", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if _, err := f.Seek(0, 0); err != nil {
		logger.Get().Warn("jsonlToCSV: failed to rewind source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	scanner = bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		v, err := p.Parse(line)
		if err != nil {
			continue
		}
		o, err := v.Object()
		if err != nil || o == nil {
			continue
		}

		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = formatValueForCSVFromFastjson(o.Get(h))
		}
		if err := w.Write(record); err != nil {
			logger.Get().Warn("jsonlToCSV: failed to write CSV record", zap.Error(err))
			return vf.vm.ToValue(false)
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Get().Warn("jsonlToCSV: failed to read source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		logger.Get().Warn("jsonlToCSV: failed to flush CSV writer", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}

func (vf *vmFunc) csvToJSONL(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling csvToJSONL", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("csvToJSONL: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	f, err := os.Open(source)
	if err != nil {
		logger.Get().Warn("csvToJSONL: failed to open source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		logger.Get().Warn("csvToJSONL: failed to parse CSV", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	if len(records) == 0 {
		return vf.vm.ToValue(false)
	}
	if len(records[0]) == 0 {
		return vf.vm.ToValue(false)
	}

	headers := records[0]

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		logger.Get().Warn("csvToJSONL: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	var sb strings.Builder
	for _, row := range records[1:] {
		obj := make(map[string]interface{}, len(headers))
		for i, h := range headers {
			if i < len(row) {
				obj[h] = row[i]
			} else {
				obj[h] = ""
			}
		}
		b, err := json.Marshal(obj)
		if err != nil {
			continue
		}
		sb.Write(b)
		sb.WriteString("\n")
	}

	if err := os.WriteFile(dest, []byte(sb.String()), 0644); err != nil {
		logger.Get().Warn("csvToJSONL: failed to write destination", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}

func jsonlUniqueFieldsFromGojaValue(v goja.Value) []string {
	exported := v.Export()
	if exported != nil {
		switch t := exported.(type) {
		case []string:
			out := make([]string, 0, len(t))
			for _, s := range t {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
			return out
		case []interface{}:
			out := make([]string, 0, len(t))
			for _, item := range t {
				s, ok := item.(string)
				if !ok {
					continue
				}
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
			return out
		case string:
			out := make([]string, 0, 8)
			for _, part := range strings.Split(t, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					out = append(out, part)
				}
			}
			return out
		}
	}

	s := strings.TrimSpace(v.String())
	if s == "" || s == "undefined" {
		return nil
	}
	out := make([]string, 0, 8)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func getFastjsonValueByPath(v *fastjson.Value, path []string) *fastjson.Value {
	cur := v
	for _, p := range path {
		o, err := cur.Object()
		if err != nil || o == nil {
			return nil
		}
		cur = o.Get(p)
		if cur == nil {
			return nil
		}
	}
	return cur
}

func fastjsonHashValue(v *fastjson.Value) string {
	if v == nil || v.Type() == fastjson.TypeNull {
		return ""
	}
	if v.Type() == fastjson.TypeString {
		return string(v.GetStringBytes())
	}
	return string(v.MarshalTo(nil))
}

func (vf *vmFunc) jsonlFilter(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling jsonlFilter")

	if len(call.Arguments) < 3 {
		logger.Get().Warn("jsonlFilter: requires 3 arguments")
		return vf.vm.ToValue(false)
	}

	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	fields := jsonlUniqueFieldsFromGojaValue(call.Argument(2))

	logger.Get().Debug("jsonlFilter arguments", zap.String("source", source), zap.String("dest", dest), zap.Int("fields", len(fields)))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" || len(fields) == 0 {
		logger.Get().Warn("jsonlFilter: empty source/dest or fields")
		return vf.vm.ToValue(false)
	}

	f, err := os.Open(source)
	if err != nil {
		logger.Get().Warn("jsonlFilter: failed to open source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = f.Close() }()

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		logger.Get().Warn("jsonlFilter: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	out, err := os.Create(dest)
	if err != nil {
		logger.Get().Warn("jsonlFilter: failed to create destination file", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = out.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	writer := bufio.NewWriterSize(out, 256*1024)
	defer func() { _ = writer.Flush() }()

	lineCount := 0
	outCount := 0
	var p fastjson.Parser
	var arena fastjson.Arena

	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		lineCount++

		v, err := p.Parse(raw)
		if err != nil {
			continue
		}
		arena.Reset()
		filtered := arena.NewObject()
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			path := strings.Split(field, ".")
			value := getFastjsonValueByPath(v, path)
			if value == nil || value.Type() == fastjson.TypeNull {
				continue
			}
			filtered.Set(field, value)
		}

		if _, err := writer.Write(filtered.MarshalTo(nil)); err != nil {
			logger.Get().Warn("jsonlFilter: failed writing output", zap.Error(err))
			return vf.vm.ToValue(false)
		}
		if err := writer.WriteByte('\n'); err != nil {
			logger.Get().Warn("jsonlFilter: failed writing newline", zap.Error(err))
			return vf.vm.ToValue(false)
		}
		outCount++
	}

	if err := scanner.Err(); err != nil {
		logger.Get().Warn("jsonlFilter: failed reading source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	if err := writer.Flush(); err != nil {
		logger.Get().Warn("jsonlFilter: failed flushing output", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("jsonlFilter completed", zap.Int("lines", lineCount), zap.Int("written", outCount))
	return vf.vm.ToValue(true)
}

func (vf *vmFunc) jsonlUnique(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling jsonlUnique")

	if len(call.Arguments) < 3 {
		logger.Get().Warn("jsonlUnique: requires 3 arguments")
		return vf.vm.ToValue(false)
	}

	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	fields := jsonlUniqueFieldsFromGojaValue(call.Argument(2))

	logger.Get().Debug("jsonlUnique arguments", zap.String("source", source), zap.String("dest", dest), zap.Int("fields", len(fields)))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" || len(fields) == 0 {
		logger.Get().Warn("jsonlUnique: empty source/dest or fields")
		return vf.vm.ToValue(false)
	}

	f, err := os.Open(source)
	if err != nil {
		logger.Get().Warn("jsonlUnique: failed to open source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = f.Close() }()

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		logger.Get().Warn("jsonlUnique: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	out, err := os.Create(dest)
	if err != nil {
		logger.Get().Warn("jsonlUnique: failed to create destination file", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = out.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	writer := bufio.NewWriterSize(out, 256*1024)
	defer func() { _ = writer.Flush() }()

	seen := make(map[string]struct{}, 1024)
	uniqueCount := 0
	lineCount := 0
	var p fastjson.Parser

	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		lineCount++

		v, err := p.Parse(raw)
		if err != nil {
			continue
		}

		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field == "" {
				parts = append(parts, "")
				continue
			}
			path := strings.Split(field, ".")
			value := getFastjsonValueByPath(v, path)
			parts = append(parts, fastjsonHashValue(value))
		}

		h := sha1.Sum([]byte(strings.Join(parts, "-")))
		hash := hex.EncodeToString(h[:])
		if _, ok := seen[hash]; ok {
			continue
		}
		seen[hash] = struct{}{}
		uniqueCount++

		if _, err := writer.WriteString(raw); err != nil {
			logger.Get().Warn("jsonlUnique: failed writing output", zap.Error(err))
			return vf.vm.ToValue(false)
		}
		if err := writer.WriteByte('\n'); err != nil {
			logger.Get().Warn("jsonlUnique: failed writing newline", zap.Error(err))
			return vf.vm.ToValue(false)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Get().Warn("jsonlUnique: failed reading source", zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	if err := writer.Flush(); err != nil {
		logger.Get().Warn("jsonlUnique: failed flushing output", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("jsonlUnique completed", zap.Int("lines", lineCount), zap.Int("unique", uniqueCount))
	return vf.vm.ToValue(true)
}

func formatValueFromFastjson(val *fastjson.Value) string {
	if val == nil {
		return ""
	}
	switch val.Type() {
	case fastjson.TypeString:
		return strings.ReplaceAll(string(val.GetStringBytes()), "|", "\\|")
	case fastjson.TypeNull:
		return ""
	default:
		b := val.MarshalTo(nil)
		return strings.ReplaceAll(string(b), "|", "\\|")
	}
}

// osmFuncPattern matches ```osm-func ... ``` code blocks in markdown
var osmFuncPattern = regexp.MustCompile("(?s)```osm-func\n(.*?)\n```")

// osmFuncInlinePattern matches `code`{.osm-func} inline syntax (Pandoc-style)
var osmFuncInlinePattern = regexp.MustCompile("`([^`]+)`\\{\\.osm-func\\}")

// renderMarkdownReport processes markdown templates with osm-func blocks
// Usage: render_markdown_report(template_path, output_path) -> bool
// Reads template, renders {{Variables}}, executes osm-func blocks, writes output
func (vf *vmFunc) renderMarkdownReport(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling renderMarkdownReport")

	if len(call.Arguments) < 2 {
		logger.Get().Warn("renderMarkdownReport: requires 2 arguments")
		return vf.vm.ToValue(false)
	}

	templatePath := call.Argument(0).String()
	outputPath := call.Argument(1).String()

	if templatePath == "undefined" || templatePath == "" {
		logger.Get().Warn("renderMarkdownReport: template_path cannot be empty")
		return vf.vm.ToValue(false)
	}

	if outputPath == "undefined" || outputPath == "" {
		logger.Get().Warn("renderMarkdownReport: output_path cannot be empty")
		return vf.vm.ToValue(false)
	}

	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		logger.Get().Warn("renderMarkdownReport: failed to read template",
			zap.String("path", templatePath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Get context variables from the Goja VM
	ctx := vf.getContextVariables()

	// Step 1: Render {{Variable}} template variables first
	rendered := vf.renderTemplateVariables(string(content), ctx)

	// Step 2: Find and process osm-func blocks
	processed := osmFuncPattern.ReplaceAllStringFunc(rendered, func(match string) string {
		// Extract function code from the match
		code := vf.extractOsmFuncCode(match)
		if code == "" {
			return "<!-- ERROR: empty osm-func block -->"
		}

		// Execute the function code using the Goja VM
		output, execErr := vf.executeCode(code)
		if execErr != nil {
			logger.Get().Warn("renderMarkdownReport: function execution failed",
				zap.String("code", code), zap.Error(execErr))
			return fmt.Sprintf("<!-- ERROR: %v -->", execErr)
		}

		// Convert output to string
		switch v := output.(type) {
		case string:
			return v
		case nil:
			return ""
		default:
			return fmt.Sprintf("%v", v)
		}
	})

	// Step 3: Process `code`{.osm-func} inline syntax
	processed = osmFuncInlinePattern.ReplaceAllStringFunc(processed, func(match string) string {
		code := vf.extractOsmFuncInlineCode(match)
		if code == "" {
			return "<!-- ERROR: empty inline osm-func -->"
		}

		output, execErr := vf.executeCode(code)
		if execErr != nil {
			logger.Get().Warn("renderMarkdownReport: inline function execution failed",
				zap.String("code", code), zap.Error(execErr))
			return fmt.Sprintf("<!-- ERROR: %v -->", execErr)
		}

		switch v := output.(type) {
		case string:
			return v
		case nil:
			return ""
		default:
			return fmt.Sprintf("%v", v)
		}
	})

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Get().Warn("renderMarkdownReport: failed to create output directory",
			zap.String("path", outputDir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(processed), 0644); err != nil {
		logger.Get().Warn("renderMarkdownReport: failed to write output",
			zap.String("path", outputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("renderMarkdownReport completed successfully",
		zap.String("template", templatePath),
		zap.String("output", outputPath))

	return vf.vm.ToValue(true)
}

// getContextVariables retrieves context variables from the Goja VM
func (vf *vmFunc) getContextVariables() map[string]interface{} {
	ctx := make(map[string]interface{})

	// List of common workflow variables to retrieve
	varNames := []string{
		"Target", "TargetSpace", "Workspace", "Output", "BaseFolder",
		"Binaries", "Data", "ExternalConfigs", "Workflows", "Workspaces",
		"RunUUID", "TaskDate", "Version", "DefaultUA",
	}

	for _, name := range varNames {
		val := vf.vm.Get(name)
		if val != nil && !goja.IsUndefined(val) {
			exported := val.Export()
			if exported != nil {
				ctx[name] = exported
			}
		}
	}

	// Also add runtime fields
	vmCtx := vf.getContext()
	if vmCtx != nil {
		if vmCtx.workspaceName != "" {
			ctx["Workspace"] = vmCtx.workspaceName
			ctx["TargetSpace"] = vmCtx.workspaceName
		}
		if vmCtx.scanID != "" {
			ctx["RunUUID"] = vmCtx.scanID
		}
	}

	return ctx
}

// renderTemplateVariables renders {{Variable}} syntax in the content
func (vf *vmFunc) renderTemplateVariables(content string, ctx map[string]interface{}) string {
	// Pattern: {{VariableName}} where VariableName is alphanumeric/underscore
	varPattern := regexp.MustCompile(`\{\{(\w+)\}\}`)

	result := varPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name (without braces)
		varName := match[2 : len(match)-2]

		if val, ok := ctx[varName]; ok {
			return fmt.Sprintf("%v", val)
		}

		// Return original if variable not found
		return match
	})

	return result
}

// extractOsmFuncCode extracts the function code from an osm-func block
func (vf *vmFunc) extractOsmFuncCode(block string) string {
	// Remove the ```osm-func\n prefix and \n``` suffix
	code := strings.TrimPrefix(block, "```osm-func\n")
	code = strings.TrimSuffix(code, "\n```")
	return strings.TrimSpace(code)
}

// extractOsmFuncInlineCode extracts function code from inline syntax
func (vf *vmFunc) extractOsmFuncInlineCode(match string) string {
	// Remove the ` prefix and `{.osm-func} suffix
	code := strings.TrimPrefix(match, "`")
	code = strings.TrimSuffix(code, "`{.osm-func}")
	return strings.TrimSpace(code)
}

// executeCode executes JavaScript code in the Goja VM and returns the result
func (vf *vmFunc) executeCode(code string) (interface{}, error) {
	result, err := vf.vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	exported := result.Export()
	return exported, nil
}

// generateSecurityReport renders a security report template and registers it as an artifact
// Usage: generate_security_report(template_path) -> bool
// Output is written to {{Output}}/security-report.md
func (vf *vmFunc) generateSecurityReport(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("generateSecurityReport"))

	if len(call.Arguments) < 1 {
		logger.Get().Warn("generateSecurityReport: requires 1 argument")
		return vf.vm.ToValue(false)
	}

	templatePath := call.Argument(0).String()
	if templatePath == "undefined" || templatePath == "" {
		logger.Get().Warn("generateSecurityReport: template_path cannot be empty")
		return vf.vm.ToValue(false)
	}

	// Get output path from workspacePath ({{Output}})
	if vf.getContext().workspacePath == "" {
		logger.Get().Warn("generateSecurityReport: Output path not set in context")
		return vf.vm.ToValue(false)
	}
	outputPath := filepath.Join(vf.getContext().workspacePath, "security-report.md")

	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		logger.Get().Warn("generateSecurityReport: failed to read template",
			zap.String("path", templatePath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Get context variables from the Goja VM
	ctx := vf.getContextVariables()

	// Step 1: Render {{Variable}} template variables first
	rendered := vf.renderTemplateVariables(string(content), ctx)

	// Step 2: Find and process osm-func blocks
	processed := osmFuncPattern.ReplaceAllStringFunc(rendered, func(match string) string {
		code := vf.extractOsmFuncCode(match)
		if code == "" {
			return "<!-- ERROR: empty osm-func block -->"
		}

		output, execErr := vf.executeCode(code)
		if execErr != nil {
			logger.Get().Warn("generateSecurityReport: function execution failed",
				zap.String("code", code), zap.Error(execErr))
			return fmt.Sprintf("<!-- ERROR: %v -->", execErr)
		}

		switch v := output.(type) {
		case string:
			return v
		case nil:
			return ""
		default:
			return fmt.Sprintf("%v", v)
		}
	})

	// Step 3: Process `code`{.osm-func} inline syntax
	processed = osmFuncInlinePattern.ReplaceAllStringFunc(processed, func(match string) string {
		code := vf.extractOsmFuncInlineCode(match)
		if code == "" {
			return "<!-- ERROR: empty inline osm-func -->"
		}

		output, execErr := vf.executeCode(code)
		if execErr != nil {
			logger.Get().Warn("generateSecurityReport: inline function execution failed",
				zap.String("code", code), zap.Error(execErr))
			return fmt.Sprintf("<!-- ERROR: %v -->", execErr)
		}

		switch v := output.(type) {
		case string:
			return v
		case nil:
			return ""
		default:
			return fmt.Sprintf("%v", v)
		}
	})

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Get().Warn("generateSecurityReport: failed to create output directory",
			zap.String("path", outputDir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(processed), 0644); err != nil {
		logger.Get().Warn("generateSecurityReport: failed to write output",
			zap.String("path", outputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Register artifact
	if err := vf.registerReportArtifact(outputPath, "security-report", "Security report summary"); err != nil {
		logger.Get().Warn("generateSecurityReport: failed to register artifact",
			zap.String("path", outputPath), zap.Error(err))
		// Don't fail the function if artifact registration fails
	}

	logger.Get().Debug(terminal.HiGreen("generateSecurityReport")+" completed successfully",
		zap.String("template", templatePath),
		zap.String("output", outputPath))

	return vf.vm.ToValue(true)
}

// registerReportArtifact registers a report file as an artifact in the database
func (vf *vmFunc) registerReportArtifact(filePath, name, description string) error {
	db := database.GetDB()
	if db == nil {
		return nil // No database, skip registration
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	lineCount := 0
	sizeBytes := info.Size()
	if !info.IsDir() {
		lineCount, _ = countNonEmptyLines(filePath)
	}

	vmCtx := vf.getContext()
	var runID int64
	workspace := ""
	if vmCtx != nil {
		runID = vmCtx.runID
		workspace = vmCtx.workspaceName
	}

	artifact := database.Artifact{
		ID:           uuid.New().String(),
		RunID:        runID,
		Workspace:    workspace,
		Name:         name,
		ArtifactPath: filePath,
		ArtifactType: database.ArtifactTypeReport,
		ContentType:  database.ContentTypeMarkdown,
		SizeBytes:    sizeBytes,
		LineCount:    lineCount,
		Description:  description,
		CreatedAt:    time.Now(),
	}

	ctx := context.Background()
	_, err = db.NewInsert().Model(&artifact).
		On("CONFLICT (id) DO UPDATE").
		Set("artifact_path = EXCLUDED.artifact_path").
		Set("size_bytes = EXCLUDED.size_bytes").
		Set("line_count = EXCLUDED.line_count").
		Set("description = EXCLUDED.description").
		Exec(ctx)

	return err
}
