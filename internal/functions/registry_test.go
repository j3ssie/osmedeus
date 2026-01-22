package functions

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	result, err := registry.Execute(
		`file_exists("`+testFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestRegistry_FileLength(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "lines.txt")
	err := os.WriteFile(testFile, []byte("line1\nline2\nline3\n"), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	result, err := registry.Execute(
		`file_length("`+testFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(3), result)
}

func TestRegistry_CreateFolder(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "nested", "dir")

	registry := NewRegistry()
	result, err := registry.Execute(
		`create_folder("`+newDir+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	info, err := os.Stat(newDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestRegistry_AppendFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.txt")
	dest := filepath.Join(tmpDir, "dest.txt")

	require.NoError(t, os.WriteFile(src, []byte("b\n"), 0644))
	require.NoError(t, os.WriteFile(dest, []byte("a\n"), 0644))

	registry := NewRegistry()
	result, err := registry.Execute(
		`append_file("`+dest+`", "`+src+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	content, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "a\nb\n", string(content))
}

func TestRegistry_GrepHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "in.txt")
	require.NoError(t, os.WriteFile(source, []byte("admin\nuser\nmyadmin\n"), 0644))

	dest1 := filepath.Join(tmpDir, "out_string.txt")
	dest2 := filepath.Join(tmpDir, "out_regex.txt")

	registry := NewRegistry()

	result, err := registry.Execute(
		`grep_string_to_file("`+dest1+`", "`+source+`", "admin")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	out1, err := os.ReadFile(dest1)
	require.NoError(t, err)
	assert.Equal(t, "admin\nmyadmin\n", string(out1))

	result, err = registry.Execute(
		`grep_regex_to_file("`+dest2+`", "`+source+`", "^a")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	out2, err := os.ReadFile(dest2)
	require.NoError(t, err)
	assert.Equal(t, "admin\n", string(out2))

	str, err := registry.Execute(
		`grep_string("`+source+`", "admin")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, "admin\nmyadmin", str)

	str, err = registry.Execute(
		`grep_regex("`+source+`", "^a")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, "admin", str)
}

func TestRegistry_Glob(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "a.txt")
	file2 := filepath.Join(tmpDir, "b.txt")
	file3 := filepath.Join(tmpDir, "c.log")
	require.NoError(t, os.WriteFile(file1, []byte("a"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("b"), 0644))
	require.NoError(t, os.WriteFile(file3, []byte("c"), 0644))

	registry := NewRegistry()
	result, err := registry.Execute(
		`glob("`+filepath.Join(tmpDir, "*.txt")+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	switch v := result.(type) {
	case []string:
		require.Len(t, v, 2)
		assert.Equal(t, file1, v[0])
		assert.Equal(t, file2, v[1])
	case []interface{}:
		require.Len(t, v, 2)
		assert.Equal(t, file1, v[0])
		assert.Equal(t, file2, v[1])
	default:
		require.Fail(t, "unexpected result type")
	}
}

func TestRegistry_JQFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "data.json")
	require.NoError(t, os.WriteFile(jsonFile, []byte(`{"url":"example.com"}`), 0644))

	registry := NewRegistry()
	result, err := registry.Execute(
		`jq_from_file("`+jsonFile+`", ".url")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, "example.com", result)
}

func TestRegistry_JSONLCSVConversion(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlFile := filepath.Join(tmpDir, "in.jsonl")
	csvFile := filepath.Join(tmpDir, "out.csv")
	jsonlOut := filepath.Join(tmpDir, "out.jsonl")

	content := "{\"name\":\"Alice\",\"age\":30}\n{\"name\":\"Bob\"}\n"
	require.NoError(t, os.WriteFile(jsonlFile, []byte(content), 0644))

	registry := NewRegistry()
	result, err := registry.Execute(
		`jsonl_to_csv("`+jsonlFile+`", "`+csvFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	csvContent, err := os.ReadFile(csvFile)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(csvContent)), "\n")
	require.GreaterOrEqual(t, len(lines), 2)
	// Columns should be in the order they appear in the first JSON object
	assert.Equal(t, "name,age", lines[0])

	result, err = registry.Execute(
		`csv_to_jsonl("`+csvFile+`", "`+jsonlOut+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	jsonlBytes, err := os.ReadFile(jsonlOut)
	require.NoError(t, err)
	outLines := strings.Split(strings.TrimSpace(string(jsonlBytes)), "\n")
	require.Len(t, outLines, 2)

	var row1 map[string]interface{}
	var row2 map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(outLines[0]), &row1))
	require.NoError(t, json.Unmarshal([]byte(outLines[1]), &row2))
	assert.Equal(t, "Alice", row1["name"])
	assert.Equal(t, "30", row1["age"])
	assert.Equal(t, "Bob", row2["name"])
}

func TestRegistry_JSONLFilter(t *testing.T) {
	tmpDir := t.TempDir()
	in := filepath.Join(tmpDir, "in.jsonl")
	out := filepath.Join(tmpDir, "out.jsonl")

	content := "{\"name\":\"Alice\",\"age\":30,\"hash\":{\"body_sha256\":\"abc\"}}\n" +
		"{\"name\":\"Bob\",\"age\":25}\n"
	require.NoError(t, os.WriteFile(in, []byte(content), 0644))

	registry := NewRegistry()
	result, err := registry.Execute(
		"jsonl_filter(\""+in+"\", \""+out+"\", 'name,hash.body_sha256')",
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	bytes, err := os.ReadFile(out)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(bytes)), "\n")
	require.Len(t, lines, 2)

	var row1 map[string]interface{}
	var row2 map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &row1))
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &row2))

	assert.Equal(t, "Alice", row1["name"])
	assert.Equal(t, "abc", row1["hash.body_sha256"])
	assert.Equal(t, "Bob", row2["name"])
	_, hasNested := row2["hash.body_sha256"]
	assert.False(t, hasNested)
}

func TestRegistry_Trim(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`trim("  hello  ")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestRegistry_Contains(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`contains("hello world", "world")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestRegistry_EvaluateCondition(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.EvaluateCondition(
		`fileLength > 0`,
		map[string]interface{}{"fileLength": 10},
	)

	require.NoError(t, err)
	assert.True(t, result)
}

func TestRegistry_StoreArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")
	cfg := &config.Config{
		BaseFolder: tmpDir,
		Database: config.DatabaseConfig{
			DBEngine: "sqlite",
			DBPath:   dbPath,
		},
	}

	_, err := database.Connect(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = database.Close()
		database.SetDB(nil)
	})

	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx))

	filePath := filepath.Join(tmpDir, "out.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("a\n\n b\n"), 0644))

	registry := NewRegistry()

	result, err := registry.Execute(
		`store_artifact("`+filePath+`")`,
		map[string]interface{}{
			"Workspace": "w1",
			"RunUUID":   "r1",
			"DBRunID":   int64(1),
		},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	db := database.GetDB()
	require.NotNil(t, db)

	var artifacts []database.Artifact
	err = db.NewSelect().Model(&artifacts).
		Where("run_id = ?", int64(1)).
		Where("workspace = ?", "w1").
		Scan(ctx)
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	assert.Equal(t, filePath, artifacts[0].ArtifactPath)
	assert.Equal(t, "out.txt", artifacts[0].Name)
}
