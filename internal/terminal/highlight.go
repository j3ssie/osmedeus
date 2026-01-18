package terminal

import (
	"bytes"
	"io"
	"os"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Highlighter provides syntax highlighting for terminal output
type Highlighter struct {
	style     *chroma.Style
	formatter chroma.Formatter
}

// NewHighlighter creates a new syntax highlighter
func NewHighlighter() *Highlighter {
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	var formatter chroma.Formatter
	if colorEnabled {
		formatter = formatters.Get("terminal256")
		if formatter == nil {
			formatter = formatters.Get("terminal")
		}
	} else {
		formatter = formatters.Get("noop")
	}

	if formatter == nil {
		formatter = formatters.Fallback
	}

	return &Highlighter{
		style:     style,
		formatter: formatter,
	}
}

// highlight performs syntax highlighting on content
func (h *Highlighter) highlight(content, language string) (string, error) {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content, err
	}

	var buf bytes.Buffer
	err = h.formatter.Format(&buf, h.style, iterator)
	if err != nil {
		return content, err
	}

	return buf.String(), nil
}

// HighlightYAML highlights YAML content
func (h *Highlighter) HighlightYAML(content string) (string, error) {
	return h.highlight(content, "yaml")
}

// HighlightMarkdown highlights Markdown content
func (h *Highlighter) HighlightMarkdown(content string) (string, error) {
	return h.highlight(content, "markdown")
}

// HighlightBash highlights Bash/shell content
func (h *Highlighter) HighlightBash(content string) (string, error) {
	return h.highlight(content, "bash")
}

// HighlightJSON highlights JSON content
func (h *Highlighter) HighlightJSON(content string) (string, error) {
	return h.highlight(content, "json")
}

// HighlightGo highlights Go content
func (h *Highlighter) HighlightGo(content string) (string, error) {
	return h.highlight(content, "go")
}

// Highlight highlights content with auto-detected language
func (h *Highlighter) Highlight(content string) (string, error) {
	lexer := lexers.Analyse(content)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content, err
	}

	var buf bytes.Buffer
	err = h.formatter.Format(&buf, h.style, iterator)
	if err != nil {
		return content, err
	}

	return buf.String(), nil
}

// PrintHighlighted prints highlighted content to stdout
func (h *Highlighter) PrintHighlighted(content, language string) error {
	highlighted, err := h.highlight(content, language)
	if err != nil {
		// Fallback to plain output
		_, err = io.WriteString(os.Stdout, content)
		return err
	}
	_, err = io.WriteString(os.Stdout, highlighted)
	return err
}

// PrintHighlightedYAML prints highlighted YAML to stdout
func (h *Highlighter) PrintHighlightedYAML(content string) error {
	return h.PrintHighlighted(content, "yaml")
}

// PrintHighlightedMarkdown prints highlighted Markdown to stdout
func (h *Highlighter) PrintHighlightedMarkdown(content string) error {
	return h.PrintHighlighted(content, "markdown")
}
