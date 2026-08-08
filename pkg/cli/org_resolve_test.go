package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withOrgEnv isolates the org resolution inputs for one test: the --org flag, both
// environment variables, and the base folder holding the active-org file.
func withOrgEnv(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	prevFlag, prevBase := globalOrg, baseFolder
	globalOrg, baseFolder = "", tmpDir
	t.Setenv("OSMEDEUS_ORG_UUID", "")
	t.Setenv("OSMEDEUS_ORG", "")

	resetOrgResolution()
	t.Cleanup(func() {
		globalOrg, baseFolder = prevFlag, prevBase
		resetOrgResolution()
	})

	return tmpDir
}

func TestOrgRefUnsetMeansNoFilter(t *testing.T) {
	withOrgEnv(t)
	assert.Empty(t, orgRef(), "no org selected must resolve to empty so reads span all orgs")
}

func TestOrgRefPrecedence(t *testing.T) {
	tmpDir := withOrgEnv(t)

	// 4. active-org file is the lowest-priority source.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, activeOrgFileName), []byte("from-file\n"), 0o644))
	assert.Equal(t, "from-file", orgRef())

	// 3. $OSMEDEUS_ORG beats the file.
	t.Setenv("OSMEDEUS_ORG", "from-org-env")
	assert.Equal(t, "from-org-env", orgRef())

	// 2. $OSMEDEUS_ORG_UUID beats $OSMEDEUS_ORG.
	t.Setenv("OSMEDEUS_ORG_UUID", "from-uuid-env")
	assert.Equal(t, "from-uuid-env", orgRef())

	// 1. the --org flag beats everything.
	globalOrg = "from-flag"
	assert.Equal(t, "from-flag", orgRef())
}

func TestOrgRefTrimsWhitespace(t *testing.T) {
	withOrgEnv(t)

	globalOrg = "   "
	assert.Empty(t, orgRef(), "a blank flag must not count as a selection")

	globalOrg = "  acme  "
	assert.Equal(t, "acme", orgRef())
}

func TestActiveOrgWriteReadClear(t *testing.T) {
	withOrgEnv(t)

	require.NoError(t, writeActiveOrg("acme-uuid"))
	assert.Equal(t, "acme-uuid", readActiveOrg())
	assert.Equal(t, "acme-uuid", orgRef())

	require.NoError(t, clearActiveOrg())
	assert.Empty(t, readActiveOrg())
	assert.Empty(t, orgRef(), "clearing the active org returns to the all-orgs view")

	// Clearing twice is not an error.
	require.NoError(t, clearActiveOrg())
}

func TestActiveOrgFilePathUsesBaseFolder(t *testing.T) {
	tmpDir := withOrgEnv(t)
	assert.Equal(t, filepath.Join(tmpDir, activeOrgFileName), activeOrgFilePath())
}
