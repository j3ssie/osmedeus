package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupName(t *testing.T) {
	cases := []struct {
		name  string
		entry *BinaryEntry
		want  string
	}{
		{"coreutils", &BinaryEntry{ValidateCommand: "sort --help"}, "sort"},
		{"minio-client", &BinaryEntry{ValidateCommand: "mc"}, "mc"},
		{"jq", &BinaryEntry{ValidateCommand: "jq --version"}, "jq"},
		{"httpx", &BinaryEntry{}, "httpx"},
		{"nuclei", &BinaryEntry{ValidateCommand: "nuclei -version"}, "nuclei"},
		{"blank", &BinaryEntry{ValidateCommand: "   "}, "blank"},
		{"nil-entry", nil, "nil-entry"},
		// A check that runs another program on an argument names that program,
		// not the package: fall back to the key, never "installed because
		// python3 is".
		{"pymod", &BinaryEntry{ValidateCommand: "python3 -c 'import pymod'"}, "pymod"},
		{"templates", &BinaryEntry{ValidateCommand: "test -d /opt/templates"}, "templates"},
		{"marker", &BinaryEntry{ValidateCommand: "touch /tmp/x"}, "marker"},
		{"piped", &BinaryEntry{ValidateCommand: "foo --version | grep 1"}, "piped"},
		{"subshell", &BinaryEntry{ValidateCommand: "sh -c foo"}, "subshell"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, LookupName(tc.name, tc.entry))
		})
	}
}

// fakeBinDir puts executables with the given names on PATH, ahead of the rest.
func fakeBinDir(t *testing.T, names ...string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stubs are not portable to windows")
	}
	dir := t.TempDir()
	for _, n := range names {
		require.NoError(t, os.WriteFile(filepath.Join(dir, n), []byte("#!/bin/sh\nexit 0\n"), 0o755))
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// entryWithMarkerInstall is an entry whose install command, if it runs at all,
// creates marker — so a test can tell "skipped as installed" from "installed".
func entryWithMarkerInstall(validate, marker string) BinaryEntry {
	cmd := "touch " + marker
	return BinaryEntry{
		ValidateCommand:     validate,
		MultiCommandsLinux:  []string{cmd},
		MultiCommandsDarwin: []string{cmd},
	}
}

// Issue #330: coreutils ships no `coreutils` executable on any common distro, so
// looking up the registry key always failed and the installer fell through to
// `apt install coreutils` (and, without sudo, to exit 127). Its valide-command
// names `sort`, which is present.
func TestInstallBinary_SkipsPackageFoundByValidateCommand(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "install-ran")
	binDir := t.TempDir()
	registry := BinaryRegistry{"coreutils": entryWithMarkerInstall("sort --help", marker)}

	skipped, err := InstallBinary("coreutils", registry, binDir, nil)
	require.NoError(t, err)
	assert.True(t, skipped, "reported as skipped, not installed")

	assert.NoFileExists(t, marker, "coreutils is present (sort is on PATH); nothing should be installed")
	assert.NoFileExists(t, filepath.Join(binDir, "coreutils"),
		"a core tool must not be linked in under the package's name")
}

func TestInstallBinary_LinksNonCoreToolFoundByValidateCommand(t *testing.T) {
	fakeBinDir(t, "fakemc")
	marker := filepath.Join(t.TempDir(), "install-ran")
	binDir := t.TempDir()
	registry := BinaryRegistry{"fake-minio-client": entryWithMarkerInstall("fakemc", marker)}

	skipped, err := InstallBinary("fake-minio-client", registry, binDir, nil)
	require.NoError(t, err)
	assert.True(t, skipped)

	assert.NoFileExists(t, marker, "already installed as fakemc; must not reinstall")
	assert.FileExists(t, filepath.Join(binDir, "fake-minio-client"))
}

// A multi-word valide-command used to reach exec.LookPath whole
// (LookPath("jq --version")), which can never succeed, so the binary was never
// linked into the binaries folder.
func TestSymlinkInstalledBinaryToFolder_MultiWordValidateCommand(t *testing.T) {
	fakeBinDir(t, "fakejq")
	binDir := t.TempDir()

	entry := &BinaryEntry{ValidateCommand: "fakejq --version"}
	require.NoError(t, SymlinkInstalledBinaryToFolder("fakejq", entry, binDir))

	assert.FileExists(t, filepath.Join(binDir, "fakejq"))
}

func TestLinkIfPresent_Missing(t *testing.T) {
	binDir := t.TempDir()
	assert.False(t, LinkIfPresent("definitely-not-installed-xyz", nil, binDir))
	assert.False(t, LinkIfPresent("definitely-not-installed-xyz", &BinaryEntry{}, binDir))
}

// An interpreter-style check must not make the package look installed, nor
// link the interpreter in under the package's name.
func TestInstallBinary_InterpreterValidateCommandStillInstalls(t *testing.T) {
	fakeBinDir(t, "fakepython")
	marker := filepath.Join(t.TempDir(), "install-ran")
	binDir := t.TempDir()
	registry := BinaryRegistry{"pymod": entryWithMarkerInstall("fakepython -c 'import pymod'", marker)}

	skipped, err := InstallBinary("pymod", registry, binDir, nil)
	require.NoError(t, err)
	assert.False(t, skipped, "pymod was installed, not skipped")

	assert.FileExists(t, marker, "pymod is not on PATH; it must be installed")
	assert.NoFileExists(t, filepath.Join(binDir, "pymod"), "fakepython must not be linked as pymod")
}
