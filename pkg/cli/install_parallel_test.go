package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/installer"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #330's summary read "Installed 25 binaries (24 skipped, 1 failed)" for a
// 26-binary registry: everything already on PATH was counted as installed and
// again as skipped, and the skipped count looked binaries up by registry key
// (so coreutils was never "skipped"). installBinariesParallel now reports what
// its own pre-check skipped, so the summary has one source.
func TestInstallBinariesParallel_ReportsSkippedAndFailed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script stubs are not portable to windows")
	}
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fakepresent"), []byte("#!/bin/sh\nexit 0\n"), 0o755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("OSMEDEUS_SILENT", "1")

	fail := []string{"false"}
	registry := installer.BinaryRegistry{
		"fake-present":   {ValidateCommand: "fakepresent --version"},
		"coreutils":      {ValidateCommand: "sort --help", MultiCommandsLinux: fail, MultiCommandsDarwin: fail},
		"fake-missing-x": {MultiCommandsLinux: fail, MultiCommandsDarwin: fail},
	}

	skipped, failed := installBinariesParallel(
		[]string{"coreutils", "fake-missing-x", "fake-present"},
		registry, t.TempDir(), nil, terminal.NewPrinter(), true)

	assert.ElementsMatch(t, []string{"coreutils", "fake-present"}, skipped)
	assert.Equal(t, []string{"fake-missing-x"}, failed)
}

func TestInstallBinariesParallel_AllPresent(t *testing.T) {
	registry := installer.BinaryRegistry{"coreutils": {ValidateCommand: "sort --help"}}
	skipped, failed := installBinariesParallel([]string{"coreutils"}, registry, t.TempDir(), nil, terminal.NewPrinter(), true)
	assert.Equal(t, []string{"coreutils"}, skipped, "the all-present early return reports them as skipped")
	assert.Empty(t, failed)
}
