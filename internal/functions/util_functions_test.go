package functions

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecCmd(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("simple echo command", func(t *testing.T) {
		result, err := runtime.Execute(`exec_cmd("echo hello")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("command with pipe", func(t *testing.T) {
		result, err := runtime.Execute(`exec_cmd("echo 'hello world' | cut -d' ' -f2")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "world", result)
	})

	t.Run("empty command returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`exec_cmd("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("invalid command returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`exec_cmd("nonexistentcmd12345")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("command with newlines trimmed", func(t *testing.T) {
		result, err := runtime.Execute(`exec_cmd("printf 'hello\\n'")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})
}

func TestBash(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("simple echo command", func(t *testing.T) {
		result, err := runtime.Execute(`bash("echo hello")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("empty command returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`bash("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestCutWithDelim(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("extract first field", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("a,b,c", ",", 1)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "a", result)
	})

	t.Run("extract middle field", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("a,b,c", ",", 2)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "b", result)
	})

	t.Run("extract last field", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("a,b,c", ",", 3)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "c", result)
	})

	t.Run("field out of range returns empty", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("a,b,c", ",", 5)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("field zero returns empty", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("a,b,c", ",", 0)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("negative field returns empty", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("a,b,c", ",", -1)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("different delimiter", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("a:b:c", ":", 2)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "b", result)
	})

	t.Run("multi-char delimiter", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("a::b::c", "::", 2)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "b", result)
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("", ",", 1)`, nil)
		require.NoError(t, err)
		// Empty string split by delimiter gives [""], so field 1 is ""
		assert.Equal(t, "", result)
	})

	t.Run("real-world URL parsing", func(t *testing.T) {
		result, err := runtime.Execute(`cut_with_delim("https://example.com/path", "/", 3)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "example.com", result)
	})
}

func TestLogDebug(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("logs debug message with prefix", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`log_debug("test debug message")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "[DEBUG]")
		assert.Contains(t, output, "test debug message")
	})

	t.Run("empty message still logs prefix", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`log_debug("")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "[DEBUG]")
	})
}

func TestLogInfo(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("logs info message with prefix", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`log_info("test info message")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "[INFO]")
		assert.Contains(t, output, "test info message")
	})

	t.Run("empty message still logs prefix", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`log_info("")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "[INFO]")
	})
}

func TestLogWarn(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("logs warn message with prefix", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`log_warn("test warn message")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "[WARN]")
		assert.Contains(t, output, "test warn message")
	})

	t.Run("empty message still logs prefix", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`log_warn("")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "[WARN]")
	})
}

func TestRmRF(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("remove file", func(t *testing.T) {
		tmpDir := t.TempDir()
		p := filepath.Join(tmpDir, "a.txt")
		err := os.WriteFile(p, []byte("x"), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`rm_rf("`+p+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)
		_, statErr := os.Stat(p)
		assert.Error(t, statErr)
	})

	t.Run("remove folder", func(t *testing.T) {
		tmpDir := t.TempDir()
		folder := filepath.Join(tmpDir, "dir")
		err := os.MkdirAll(folder, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(folder, "a.txt"), []byte("x"), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`rm_rf("`+folder+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)
		_, statErr := os.Stat(folder)
		assert.Error(t, statErr)
	})
}

func TestRemoveAllExcept(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("keep root file", func(t *testing.T) {
		root := t.TempDir()
		keep := filepath.Join(root, "keep.txt")
		require.NoError(t, os.WriteFile(keep, []byte("keep"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("a"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(root, "sub", "b.txt"), []byte("b"), 0644))

		result, err := runtime.Execute(`remove_all_except("`+root+`", "keep.txt")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		_, err = os.Stat(keep)
		assert.NoError(t, err)
		_, err = os.Stat(filepath.Join(root, "a.txt"))
		assert.Error(t, err)
		_, err = os.Stat(filepath.Join(root, "sub"))
		assert.Error(t, err)
	})

	t.Run("keep nested file", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0755))
		keep := filepath.Join(root, "sub", "keep.txt")
		require.NoError(t, os.WriteFile(keep, []byte("keep"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(root, "sub", "remove.txt"), []byte("x"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(root, "other"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(root, "other", "x.txt"), []byte("x"), 0644))

		result, err := runtime.Execute(`remove_all_except("`+root+`", "sub/keep.txt")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		_, err = os.Stat(keep)
		assert.NoError(t, err)
		_, err = os.Stat(filepath.Join(root, "sub", "remove.txt"))
		assert.Error(t, err)
		_, err = os.Stat(filepath.Join(root, "other"))
		assert.Error(t, err)
	})
}

func TestLogError(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("logs error message with prefix", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`log_error("test error message")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "[ERROR]")
		assert.Contains(t, output, "test error message")
	})

	t.Run("empty message still logs prefix", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_, err := runtime.Execute(`log_error("")`, nil)
		require.NoError(t, err)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "[ERROR]")
	})
}

func TestPrintGreen(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("prints green message", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		result, err := runtime.Execute(`print_green("success message")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "success message", result)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "success message")
	})

	t.Run("returns the message", func(t *testing.T) {
		result, err := runtime.Execute(`print_green("test")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		result, err := runtime.Execute(`print_green("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestPrintBlue(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("prints blue message", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		result, err := runtime.Execute(`print_blue("info message")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "info message", result)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "info message")
	})

	t.Run("returns the message", func(t *testing.T) {
		result, err := runtime.Execute(`print_blue("test")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

func TestPrintYellow(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("prints yellow message", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		result, err := runtime.Execute(`print_yellow("warning message")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "warning message", result)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "warning message")
	})

	t.Run("returns the message", func(t *testing.T) {
		result, err := runtime.Execute(`print_yellow("test")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

func TestPrintRed(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("prints red message", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		result, err := runtime.Execute(`print_red("error message")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "error message", result)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "error message")
	})

	t.Run("returns the message", func(t *testing.T) {
		result, err := runtime.Execute(`print_red("test")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

func TestSetGetVar(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("set and get variable", func(t *testing.T) {
		// Set a variable
		result, err := runtime.Execute(`set_var("my_var", "hello world")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)

		// Get the variable - need to use same VM context, so chain in single expression
		result, err = runtime.Execute(`set_var("test_key", "test_value"); get_var("test_key")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "test_value", result)
	})

	t.Run("get non-existent variable returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`get_var("nonexistent_var_xyz")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("set_var with empty name returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`set_var("", "value")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("get_var with empty name returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`get_var("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("variable is available in same execution via VM", func(t *testing.T) {
		// When set_var is called, it also sets the value on the VM
		// so it can be accessed directly as a variable
		result, err := runtime.Execute(`set_var("direct_access", "direct_value"); direct_access`, nil)
		require.NoError(t, err)
		assert.Equal(t, "direct_value", result)
	})

	t.Run("set_var overwrites existing variable", func(t *testing.T) {
		result, err := runtime.Execute(`set_var("overwrite_test", "first"); set_var("overwrite_test", "second"); get_var("overwrite_test")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "second", result)
	})

	t.Run("set_var with undefined value sets empty string", func(t *testing.T) {
		result, err := runtime.Execute(`set_var("undef_test", undefined); get_var("undef_test")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestPickValid(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("returns first valid string", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid("", "", "hello")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("skips false and empty string", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid(false, "", "world")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "world", result)
	})

	t.Run("returns number when valid", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid("", false, 123)`, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(123), result)
	})

	t.Run("returns first valid when multiple valid", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid("first", "second")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "first", result)
	})

	t.Run("returns empty string when all invalid", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid("", "", "")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("returns empty string with no arguments", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid()`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("skips undefined string value", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid("undefined", "valid")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "valid", result)
	})

	t.Run("skips null value", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid(null, "valid")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "valid", result)
	})

	t.Run("skips undefined value", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid(undefined, "valid")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "valid", result)
	})

	t.Run("skips empty array", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid([], "valid")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "valid", result)
	})

	t.Run("skips empty object", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid({}, "valid")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "valid", result)
	})

	t.Run("returns true boolean", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid(false, true, "string")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("returns non-empty array", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid([], [1, 2, 3])`, nil)
		require.NoError(t, err)
		arr, ok := result.([]interface{})
		require.True(t, ok)
		assert.Len(t, arr, 3)
	})

	t.Run("returns non-empty object", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid({}, {key: "value"})`, nil)
		require.NoError(t, err)
		obj, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "value", obj["key"])
	})

	t.Run("returns zero as valid number", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid("", 0, "fallback")`, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})

	t.Run("handles whitespace-only string as invalid", func(t *testing.T) {
		result, err := runtime.Execute(`pick_valid("   ", "valid")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "valid", result)
	})

	t.Run("respects 10 argument limit", func(t *testing.T) {
		// Arguments 1-10 are empty, 11th is "valid" - should return empty
		result, err := runtime.Execute(`pick_valid("", "", "", "", "", "", "", "", "", "", "valid")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestParseParamsToFlags(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		result := parseParamsToFlags("")
		assert.Nil(t, result)
	})

	t.Run("undefined returns nil", func(t *testing.T) {
		result := parseParamsToFlags("undefined")
		assert.Nil(t, result)
	})

	t.Run("single param", func(t *testing.T) {
		result := parseParamsToFlags("threads=10")
		assert.Equal(t, []string{"-p", "threads=10"}, result)
	})

	t.Run("multiple params", func(t *testing.T) {
		result := parseParamsToFlags("threads=10,deep=true")
		assert.Equal(t, []string{"-p", "threads=10", "-p", "deep=true"}, result)
	})

	t.Run("params with spaces", func(t *testing.T) {
		result := parseParamsToFlags(" threads=10 , deep=true ")
		assert.Equal(t, []string{"-p", "threads=10", "-p", "deep=true"}, result)
	})

	t.Run("skips entries without equals sign", func(t *testing.T) {
		result := parseParamsToFlags("threads=10,invalid,deep=true")
		assert.Equal(t, []string{"-p", "threads=10", "-p", "deep=true"}, result)
	})

	t.Run("skips empty entries from trailing comma", func(t *testing.T) {
		result := parseParamsToFlags("threads=10,")
		assert.Equal(t, []string{"-p", "threads=10"}, result)
	})
}

func TestRunModule(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("empty module returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`run_module("", "example.com")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("empty target returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`run_module("subdomain", "")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("undefined module returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`run_module(undefined, "example.com")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("undefined target returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`run_module("subdomain", undefined)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestRunFlow(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("empty flow returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`run_flow("", "example.com")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("empty target returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`run_flow("general", "")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("undefined flow returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`run_flow(undefined, "example.com")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("undefined target returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`run_flow("general", undefined)`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestExecPython(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("simple print", func(t *testing.T) {
		result, err := runtime.Execute(`exec_python("print('hello')")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("multiline code", func(t *testing.T) {
		result, err := runtime.Execute(`exec_python("x = 2 + 3\nprint(x)")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "5", result)
	})

	t.Run("empty code returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`exec_python("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("invalid code returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`exec_python("import sys; sys.exit(1)")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestExecPythonFile(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("run temp python file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pyFile := filepath.Join(tmpDir, "test.py")
		err := os.WriteFile(pyFile, []byte("print('from file')"), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`exec_python_file("`+pyFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "from file", result)
	})

	t.Run("empty path returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`exec_python_file("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("nonexistent file returns empty string", func(t *testing.T) {
		result, err := runtime.Execute(`exec_python_file("/tmp/nonexistent_py_file_12345.py")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestMoveFile(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("move file within same directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")

		content := "test content for move"
		err := os.WriteFile(source, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`move_file("`+source+`", "`+dest+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Verify source no longer exists
		_, statErr := os.Stat(source)
		assert.True(t, os.IsNotExist(statErr))

		// Verify dest exists with correct content
		destContent, err := os.ReadFile(dest)
		require.NoError(t, err)
		assert.Equal(t, content, string(destContent))
	})

	t.Run("move file to new directory (creates directory)", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "subdir", "nested", "dest.txt")

		content := "test content for nested move"
		err := os.WriteFile(source, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`move_file("`+source+`", "`+dest+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Verify source no longer exists
		_, statErr := os.Stat(source)
		assert.True(t, os.IsNotExist(statErr))

		// Verify dest exists with correct content
		destContent, err := os.ReadFile(dest)
		require.NoError(t, err)
		assert.Equal(t, content, string(destContent))
	})

	t.Run("move non-existent source fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "nonexistent.txt")
		dest := filepath.Join(tmpDir, "dest.txt")

		result, err := runtime.Execute(`move_file("`+source+`", "`+dest+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("move directory fails (file only)", func(t *testing.T) {
		tmpDir := t.TempDir()
		sourceDir := filepath.Join(tmpDir, "sourcedir")
		dest := filepath.Join(tmpDir, "dest")

		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)

		result, err := runtime.Execute(`move_file("`+sourceDir+`", "`+dest+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Verify source directory still exists
		info, statErr := os.Stat(sourceDir)
		require.NoError(t, statErr)
		assert.True(t, info.IsDir())
	})

	t.Run("empty source argument fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		dest := filepath.Join(tmpDir, "dest.txt")

		result, err := runtime.Execute(`move_file("", "`+dest+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("empty dest argument fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")

		err := os.WriteFile(source, []byte("content"), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`move_file("`+source+`", "")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Verify source still exists
		_, statErr := os.Stat(source)
		require.NoError(t, statErr)
	})

	t.Run("preserves file permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")

		content := "test content"
		err := os.WriteFile(source, []byte(content), 0755)
		require.NoError(t, err)

		result, err := runtime.Execute(`move_file("`+source+`", "`+dest+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Verify dest has executable permission
		info, statErr := os.Stat(dest)
		require.NoError(t, statErr)
		// Check that execute bit is set (at least for owner)
		assert.True(t, info.Mode()&0100 != 0, "expected executable permission to be preserved")
	})
}

func TestSkip(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("skip without message returns ErrSkipModule", func(t *testing.T) {
		_, err := runtime.Execute(`skip()`, nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrSkipModule), "error should wrap ErrSkipModule")
	})

	t.Run("skip with message preserves message", func(t *testing.T) {
		_, err := runtime.Execute(`skip('target not applicable')`, nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrSkipModule), "error should wrap ErrSkipModule")
		var skipErr *SkipModuleError
		assert.True(t, errors.As(err, &skipErr), "error should be extractable as SkipModuleError")
		assert.Equal(t, "target not applicable", skipErr.Message)
	})

	t.Run("skip default message", func(t *testing.T) {
		_, err := runtime.Execute(`skip()`, nil)
		require.Error(t, err)
		var skipErr *SkipModuleError
		if errors.As(err, &skipErr) {
			assert.Equal(t, "skip() called", skipErr.Message)
		}
	})
}

func TestSkipModuleError(t *testing.T) {
	t.Run("Error returns formatted message", func(t *testing.T) {
		err := &SkipModuleError{Message: "test message"}
		assert.Equal(t, "skip module: test message", err.Error())
	})

	t.Run("Unwrap returns ErrSkipModule", func(t *testing.T) {
		err := &SkipModuleError{Message: "test"}
		assert.Equal(t, ErrSkipModule, err.Unwrap())
	})

	t.Run("errors.Is works through chain", func(t *testing.T) {
		skipErr := &SkipModuleError{Message: "inner"}
		wrapped := errors.Join(errors.New("outer"), skipErr)
		assert.True(t, errors.Is(wrapped, ErrSkipModule))
	})
}
