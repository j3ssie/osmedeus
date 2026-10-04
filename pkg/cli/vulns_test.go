package cli

import (
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetVulnSeverityFlags clears the package-level flags severitySelection reads, so
// each case starts from the same state regardless of order.
func resetVulnSeverityFlags() {
	vulnsSeverity = ""
	vulnsMinSeverity = ""
}

func TestSeveritySelection(t *testing.T) {
	t.Run("no flags means no filtering", func(t *testing.T) {
		resetVulnSeverityFlags()
		got, err := severitySelection()
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("single severity", func(t *testing.T) {
		resetVulnSeverityFlags()
		vulnsSeverity = "high"
		got, err := severitySelection()
		require.NoError(t, err)
		assert.Equal(t, []string{"high"}, got)
	})

	t.Run("severity list is normalised and trimmed", func(t *testing.T) {
		resetVulnSeverityFlags()
		vulnsSeverity = " Critical , HIGH "
		got, err := severitySelection()
		require.NoError(t, err)
		assert.Equal(t, []string{"critical", "high"}, got)
	})

	t.Run("min-severity expands to everything at or above", func(t *testing.T) {
		resetVulnSeverityFlags()
		vulnsMinSeverity = "medium"
		got, err := severitySelection()
		require.NoError(t, err)
		// highest first, and never includes the lower bands
		assert.Equal(t, []string{"critical", "high", "medium"}, got)
	})

	t.Run("min-severity info includes every real band", func(t *testing.T) {
		resetVulnSeverityFlags()
		vulnsMinSeverity = "info"
		got, err := severitySelection()
		require.NoError(t, err)
		assert.Equal(t, []string{"critical", "high", "medium", "low", "info"}, got)
		// "unknown" is an internal alias for info, not a user-facing band
		assert.NotContains(t, got, "unknown")
	})

	t.Run("invalid severity is rejected", func(t *testing.T) {
		resetVulnSeverityFlags()
		vulnsSeverity = "bogus"
		_, err := severitySelection()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid severity")
	})

	t.Run("invalid min-severity is rejected", func(t *testing.T) {
		resetVulnSeverityFlags()
		vulnsMinSeverity = "bogus"
		_, err := severitySelection()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid --min-severity")
	})

	t.Run("the two flags are mutually exclusive", func(t *testing.T) {
		resetVulnSeverityFlags()
		vulnsSeverity = "high"
		vulnsMinSeverity = "low"
		_, err := severitySelection()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only one of")
	})

	resetVulnSeverityFlags()
}

func TestNormalizeSeverity(t *testing.T) {
	assert.Equal(t, "high", normalizeSeverity(" HIGH "))
	assert.Equal(t, "unknown", normalizeSeverity(""))
	assert.Equal(t, "unknown", normalizeSeverity("   "))
	assert.Equal(t, "critical", normalizeSeverity("Critical"))
}

func TestSortBySeverity(t *testing.T) {
	rows := []database.Vulnerability{
		{VulnTitle: "a", Severity: "info"},
		{VulnTitle: "b", Severity: "critical"},
		{VulnTitle: "c", Severity: "medium"},
		{VulnTitle: "d", Severity: ""}, // unknown ranks with info
		{VulnTitle: "e", Severity: "high"},
	}
	sortBySeverity(rows)

	got := make([]string, 0, len(rows))
	for _, r := range rows {
		got = append(got, r.VulnTitle)
	}
	// most severe first; the two lowest keep their input order (stable sort)
	assert.Equal(t, []string{"b", "e", "c", "a", "d"}, got)
}

func TestFilterBySeverity(t *testing.T) {
	rows := []database.Vulnerability{
		{VulnTitle: "a", Severity: "info"},
		{VulnTitle: "b", Severity: "critical"},
		{VulnTitle: "c", Severity: "MEDIUM"}, // case-insensitive
		{VulnTitle: "d", Severity: ""},
	}

	got := filterBySeverity(rows, []string{"critical", "medium"})
	require.Len(t, got, 2)
	assert.Equal(t, "b", got[0].VulnTitle)
	assert.Equal(t, "c", got[1].VulnTitle)

	t.Run("empty selection keeps nothing", func(t *testing.T) {
		assert.Empty(t, filterBySeverity(rows, nil))
	})
}
