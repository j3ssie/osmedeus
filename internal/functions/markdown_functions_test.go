package functions

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMarkdownFromFile(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("valid markdown file", func(t *testing.T) {
		tmpDir := t.TempDir()
		mdFile := filepath.Join(tmpDir, "test.md")
		content := "# Hello World\n\nThis is a **test**."
		err := os.WriteFile(mdFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`render_markdown_from_file("`+mdFile+`")`, nil)
		require.NoError(t, err)

		resultStr, ok := result.(string)
		require.True(t, ok)
		// Rendered markdown should contain the text (may have ANSI codes)
		assert.Contains(t, resultStr, "Hello World")
	})

	t.Run("empty path returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`render_markdown_from_file("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("non-existent file returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`render_markdown_from_file("/nonexistent/file.md")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestPrintMarkdownFromFile(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("valid file prints to stdout", func(t *testing.T) {
		tmpDir := t.TempDir()
		mdFile := filepath.Join(tmpDir, "test.md")
		content := "# Test Header\n\nSome content here."
		err := os.WriteFile(mdFile, []byte(content), 0644)
		require.NoError(t, err)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err = runtime.Execute(`print_markdown_from_file("`+mdFile+`")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "Test Header")
	})

	t.Run("empty path produces no output", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`print_markdown_from_file("")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		assert.Empty(t, buf.String())
	})

	t.Run("non-existent file produces no output", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`print_markdown_from_file("/nonexistent/file.md")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		assert.Empty(t, buf.String())
	})
}

func TestConvertJSONLToMarkdown(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("valid JSONL file", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonlFile := filepath.Join(tmpDir, "data.jsonl")
		outputFile := filepath.Join(tmpDir, "output.md")
		content := `{"name":"Alice","age":30}
{"name":"Bob","age":25}
{"name":"Charlie","age":35}`
		err := os.WriteFile(jsonlFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_jsonl_to_markdown("`+jsonlFile+`", "`+outputFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Read output file and verify content
		outputContent, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		resultStr := string(outputContent)

		// Check table structure - columns should be in the order they appear in the first JSON object
		assert.Contains(t, resultStr, "| name | age |")
		assert.Contains(t, resultStr, "| --- |")
		assert.Contains(t, resultStr, "Alice")
		assert.Contains(t, resultStr, "Bob")
		assert.Contains(t, resultStr, "Charlie")
		assert.Contains(t, resultStr, "30")
		assert.Contains(t, resultStr, "25")
		assert.Contains(t, resultStr, "35")
	})

	t.Run("empty file returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonlFile := filepath.Join(tmpDir, "empty.jsonl")
		outputFile := filepath.Join(tmpDir, "output.md")
		err := os.WriteFile(jsonlFile, []byte(""), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_jsonl_to_markdown("`+jsonlFile+`", "`+outputFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("skips invalid JSON lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonlFile := filepath.Join(tmpDir, "mixed.jsonl")
		outputFile := filepath.Join(tmpDir, "output.md")
		content := `{"name":"Alice"}
invalid json line
{"name":"Bob"}`
		err := os.WriteFile(jsonlFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_jsonl_to_markdown("`+jsonlFile+`", "`+outputFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Read output file and verify content
		outputContent, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		resultStr := string(outputContent)

		// Should have 2 data rows (Alice and Bob)
		assert.Contains(t, resultStr, "Alice")
		assert.Contains(t, resultStr, "Bob")
		lines := strings.Split(strings.TrimSpace(resultStr), "\n")
		assert.Equal(t, 4, len(lines)) // header + separator + 2 data rows
	})

	t.Run("non-existent file returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.md")
		result, err := runtime.Execute(`convert_jsonl_to_markdown("/nonexistent/file.jsonl", "`+outputFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("empty input path returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.md")
		result, err := runtime.Execute(`convert_jsonl_to_markdown("", "`+outputFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("empty output path returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonlFile := filepath.Join(tmpDir, "data.jsonl")
		err := os.WriteFile(jsonlFile, []byte(`{"name":"Alice"}`), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_jsonl_to_markdown("`+jsonlFile+`", "")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("missing arguments returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonlFile := filepath.Join(tmpDir, "data.jsonl")
		err := os.WriteFile(jsonlFile, []byte(`{"name":"Alice"}`), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_jsonl_to_markdown("`+jsonlFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("creates output directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonlFile := filepath.Join(tmpDir, "data.jsonl")
		outputFile := filepath.Join(tmpDir, "nested", "dir", "output.md")
		err := os.WriteFile(jsonlFile, []byte(`{"name":"Alice"}`), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_jsonl_to_markdown("`+jsonlFile+`", "`+outputFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Verify file was created
		_, err = os.Stat(outputFile)
		assert.NoError(t, err)
	})
}

func TestConvertCSVToMarkdown(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("valid CSV file", func(t *testing.T) {
		tmpDir := t.TempDir()
		csvFile := filepath.Join(tmpDir, "data.csv")
		content := `name,age,city
Alice,30,NYC
Bob,25,LA
Charlie,35,Chicago`
		err := os.WriteFile(csvFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_csv_to_markdown("`+csvFile+`")`, nil)
		require.NoError(t, err)

		resultStr, ok := result.(string)
		require.True(t, ok)

		// Check table structure
		assert.Contains(t, resultStr, "| name | age | city |")
		assert.Contains(t, resultStr, "| --- |")
		assert.Contains(t, resultStr, "Alice")
		assert.Contains(t, resultStr, "NYC")
		assert.Contains(t, resultStr, "Bob")
		assert.Contains(t, resultStr, "LA")
	})

	t.Run("empty file returns empty string", func(t *testing.T) {
		tmpDir := t.TempDir()
		csvFile := filepath.Join(tmpDir, "empty.csv")
		err := os.WriteFile(csvFile, []byte(""), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_csv_to_markdown("`+csvFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("non-existent file returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`convert_csv_to_markdown("/nonexistent/file.csv")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("empty path returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`convert_csv_to_markdown("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("CSV with pipe characters are escaped", func(t *testing.T) {
		tmpDir := t.TempDir()
		csvFile := filepath.Join(tmpDir, "pipes.csv")
		content := `name,command
test,echo hello | grep hello`
		err := os.WriteFile(csvFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_csv_to_markdown("`+csvFile+`")`, nil)
		require.NoError(t, err)

		resultStr, ok := result.(string)
		require.True(t, ok)

		// Pipe should be escaped
		assert.Contains(t, resultStr, `\|`)
	})

	t.Run("headers only returns just header row", func(t *testing.T) {
		tmpDir := t.TempDir()
		csvFile := filepath.Join(tmpDir, "headers.csv")
		content := `name,age,city`
		err := os.WriteFile(csvFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`convert_csv_to_markdown("`+csvFile+`")`, nil)
		require.NoError(t, err)

		resultStr, ok := result.(string)
		require.True(t, ok)

		lines := strings.Split(strings.TrimSpace(resultStr), "\n")
		assert.Equal(t, 2, len(lines)) // header + separator only
	})
}

func TestRenderMarkdownReport(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("basic template with variables", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "# Report for {{Workspace}}\nTarget: {{Target}}"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		ctx := map[string]interface{}{
			"Workspace": "test-workspace",
			"Target":    "example.com",
		}

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			ctx,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Verify output file
		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "test-workspace")
		assert.Contains(t, string(content), "example.com")
	})

	t.Run("osm-func blocks are executed", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		// Use backtick-style string for template with code blocks
		template := "# Test\n\n" +
			"```osm-func\n" +
			"trim(\"  hello  \")\n" +
			"```\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "hello")
		assert.NotContains(t, string(content), "osm-func")
	})

	t.Run("empty template path returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output.md")

		result, err := runtime.Execute(
			`render_markdown_report("", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("empty output path returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		err := os.WriteFile(templatePath, []byte("# Test"), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("non-existent template returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output.md")

		result, err := runtime.Execute(
			`render_markdown_report("/nonexistent/template.md", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("invalid osm-func produces error comment", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "# Test\n\n" +
			"```osm-func\n" +
			"invalid_syntax(((\n" +
			"```\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result) // Still succeeds

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "<!-- ERROR:")
	})

	t.Run("creates output directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "nested", "dir", "output.md")

		err := os.WriteFile(templatePath, []byte("# Test"), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Verify file was created
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)
	})

	t.Run("multiple osm-func blocks", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "# Test\n\n" +
			"First: ```osm-func\nlen(\"hello\")\n```\n\n" +
			"Second: ```osm-func\nlen(\"world!\")\n```\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "5") // len("hello")
		assert.Contains(t, string(content), "6") // len("world!")
	})

	t.Run("mixed variables and osm-func blocks", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "# Report for {{Workspace}}\n\n" +
			"Result: ```osm-func\ntoUpperCase(\"test\")\n```\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		ctx := map[string]interface{}{
			"Workspace": "my-workspace",
		}

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			ctx,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "my-workspace")
		assert.Contains(t, string(content), "TEST")
	})

	t.Run("osm-func with string concatenation", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "# Test\n\n" +
			"```osm-func\n" +
			"\"Hello \" + \"World\"\n" +
			"```\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Hello World")
	})

	t.Run("inline osm-func syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "Result: `len(\"hello\")`{.osm-func} chars\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Result: 5 chars")
		assert.NotContains(t, string(content), "{.osm-func}")
	})

	t.Run("multiple inline osm-func in one line", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "A: `len(\"a\")`{.osm-func} B: `len(\"bb\")`{.osm-func}\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "A: 1 B: 2")
	})

	t.Run("mixed code blocks and inline osm-func", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "Block: ```osm-func\nlen(\"abc\")\n```\n\nInline: `len(\"de\")`{.osm-func}\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Block: 3")
		assert.Contains(t, string(content), "Inline: 2")
	})

	t.Run("invalid inline osm-func produces error comment", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "template.md")
		outputPath := filepath.Join(tmpDir, "output.md")

		template := "Result: `invalid_function()`{.osm-func}\n"
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`render_markdown_report("`+templatePath+`", "`+outputPath+`")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "<!-- ERROR:")
	})

	t.Run("missing argument returns false", func(t *testing.T) {
		result, err := runtime.Execute(
			`render_markdown_report("/some/path")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})
}

func TestJSONLUnique(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("deduplicates by selected top-level fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "in.jsonl")
		dest := filepath.Join(tmpDir, "out.jsonl")

		content := "" +
			`{"url":"https://a","status":200,"words":10,"lines":2}` + "\n" +
			`{"url":"https://a","status":200,"words":10,"lines":2,"junk":"x"}` + "\n" +
			`{"url":"https://a","status":404,"words":10,"lines":2}` + "\n"
		err := os.WriteFile(source, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`jsonl_unique("`+source+`", "`+dest+`", ["status","words","lines"])`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		out, err := os.ReadFile(dest)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		require.Len(t, lines, 2)
		assert.Contains(t, lines[0], `"status":200`)
		assert.Contains(t, lines[1], `"status":404`)
	})

	t.Run("deduplicates by selected nested fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "in.jsonl")
		dest := filepath.Join(tmpDir, "out.jsonl")

		content := "" +
			`{"hash":{"body_sha256":"abc"},"words":10,"lines":2}` + "\n" +
			`{"hash":{"body_sha256":"abc"},"words":10,"lines":2,"extra":true}` + "\n" +
			`{"hash":{"body_sha256":"def"},"words":10,"lines":2}` + "\n"
		err := os.WriteFile(source, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`jsonl_unique("`+source+`", "`+dest+`", ["hash.body_sha256","words","lines"])`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		out, err := os.ReadFile(dest)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		require.Len(t, lines, 2)
		assert.Contains(t, string(out), `"body_sha256":"abc"`)
		assert.Contains(t, string(out), `"body_sha256":"def"`)
	})

	t.Run("accepts fields as comma-separated string", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "in.jsonl")
		dest := filepath.Join(tmpDir, "out.jsonl")

		content := "" +
			`{"status":200,"words":10,"lines":2}` + "\n" +
			`{"status":200,"words":10,"lines":2,"junk":"x"}` + "\n"
		err := os.WriteFile(source, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(
			`jsonl_unique("`+source+`", "`+dest+`", "status,words,lines")`,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		out, err := os.ReadFile(dest)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		require.Len(t, lines, 1)
	})
}
