package parser

import (
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Command Requirement Tests

func TestDependencyChecker_CheckCommands(t *testing.T) {
	checker := NewDependencyChecker()

	// Test with commands that should exist on all systems
	err := checker.CheckCommands([]string{"echo", "cat"}, "")
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckCommands_Missing(t *testing.T) {
	checker := NewDependencyChecker()

	// Test with a command that definitely doesn't exist
	err := checker.CheckCommands([]string{"echo", "nonexistent-tool-xyz-12345"}, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
	assert.Contains(t, err.Error(), "nonexistent-tool-xyz-12345")
}

func TestDependencyChecker_CheckCommands_Empty(t *testing.T) {
	checker := NewDependencyChecker()

	// Test with empty list
	err := checker.CheckCommands([]string{}, "")
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckCommands_AllMissing(t *testing.T) {
	checker := NewDependencyChecker()

	// Test with multiple missing commands
	err := checker.CheckCommands([]string{
		"nonexistent-cmd-abc",
		"nonexistent-cmd-xyz",
	}, "")

	require.Error(t, err)
	depErr, ok := err.(*DependencyError)
	require.True(t, ok)
	assert.Len(t, depErr.Missing, 2)
}

// Variable Requirement Tests

func TestDependencyChecker_CheckVariables(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeString,
			Required: true,
		},
	}

	ctx := map[string]interface{}{
		"target": "example.com",
	}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_Missing(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeString,
			Required: true,
		},
	}

	// Empty context - target is missing
	ctx := map[string]interface{}{}

	err := checker.CheckVariables(deps, ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target")
	assert.Contains(t, err.Error(), "missing")
}

func TestDependencyChecker_CheckVariables_NotRequired(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "optional_param",
			Type:     core.VarTypeString,
			Required: false,
		},
	}

	// Empty context - but param is not required
	ctx := map[string]interface{}{}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_DomainType(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeDomain,
			Required: true,
		},
	}

	// Valid domain
	ctx := map[string]interface{}{
		"target": "example.com",
	}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_InvalidDomain(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeDomain,
			Required: true,
		},
	}

	// Invalid domain format
	ctx := map[string]interface{}{
		"target": "not-a-valid-domain",
	}

	err := checker.CheckVariables(deps, ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain")
}

func TestDependencyChecker_CheckVariables_SubdomainType(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeSubdomain,
			Required: true,
		},
	}

	ctx := map[string]interface{}{
		"target": "a.example.com",
	}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_InvalidSubdomain(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeSubdomain,
			Required: true,
		},
	}

	ctx := map[string]interface{}{
		"target": "example.com",
	}

	err := checker.CheckVariables(deps, ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subdomain")
}

func TestDependencyChecker_CheckVariables_URLType(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeURL,
			Required: true,
		},
	}

	ctx := map[string]interface{}{
		"target": "https://example.com/path",
	}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_InvalidURL(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeURL,
			Required: true,
		},
	}

	ctx := map[string]interface{}{
		"target": "example.com",
	}

	err := checker.CheckVariables(deps, ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "url")
}

func TestDependencyChecker_CheckVariables_CIDRType(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeCIDR,
			Required: true,
		},
	}

	ctx := map[string]interface{}{
		"target": "10.0.0.0/8",
	}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_InvalidCIDR(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "target",
			Type:     core.VarTypeCIDR,
			Required: true,
		},
	}

	ctx := map[string]interface{}{
		"target": "10.0.0.0",
	}

	err := checker.CheckVariables(deps, ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cidr")
}

func TestDependencyChecker_CheckVariables_RepoType(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "repo",
			Type:     core.VarTypeRepo,
			Required: true,
		},
	}

	ctx := map[string]interface{}{
		"repo": "owner/project",
	}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_RepoType_GitURLs(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "repo",
			Type:     core.VarTypeRepo,
			Required: true,
		},
	}

	tests := []string{
		"https://github.com/osmedeus/assets",
		"https://github.com/osmedeus/assets.git",
		"git@github.com:osmedeus/assets.git",
		"github.com/osmedeus/assets",
		"git://github.com/osmedeus/assets.git",
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			ctx := map[string]interface{}{
				"repo": tc,
			}
			err := checker.CheckVariables(deps, ctx)
			assert.NoError(t, err)
		})
	}
}

func TestDependencyChecker_CheckVariables_NumberType(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "threads",
			Type:     core.VarTypeNumber,
			Required: true,
		},
	}

	// Valid number
	ctx := map[string]interface{}{
		"threads": "10",
	}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_InvalidNumber(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "threads",
			Type:     core.VarTypeNumber,
			Required: true,
		},
	}

	// Invalid number format
	ctx := map[string]interface{}{
		"threads": "not-a-number",
	}

	err := checker.CheckVariables(deps, ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "number")
}

func TestDependencyChecker_CheckVariables_PathType(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "output_dir",
			Type:     core.VarTypePath,
			Required: true,
		},
	}

	// Valid path
	ctx := map[string]interface{}{
		"output_dir": "/tmp/output",
	}

	err := checker.CheckVariables(deps, ctx)
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckVariables_EmptyPath(t *testing.T) {
	checker := NewDependencyChecker()

	deps := []core.VariableDep{
		{
			Name:     "output_dir",
			Type:     core.VarTypePath,
			Required: true,
		},
	}

	// Empty path
	ctx := map[string]interface{}{
		"output_dir": "",
	}

	err := checker.CheckVariables(deps, ctx)
	require.Error(t, err)
}

// CheckAll Tests

func TestDependencyChecker_CheckAll(t *testing.T) {
	checker := NewDependencyChecker()

	deps := &core.Dependencies{
		Commands: []string{"echo", "cat"},
		Variables: []core.VariableDep{
			{
				Name:     "target",
				Type:     core.VarTypeString,
				Required: true,
			},
		},
	}

	ctx := map[string]interface{}{
		"target": "test-value",
	}

	err := checker.CheckAll(deps, ctx, "")
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckAll_NilDeps(t *testing.T) {
	checker := NewDependencyChecker()

	// Should not error with nil dependencies
	err := checker.CheckAll(nil, map[string]interface{}{}, "")
	assert.NoError(t, err)
}

func TestDependencyChecker_CheckAll_CommandsFail(t *testing.T) {
	checker := NewDependencyChecker()

	deps := &core.Dependencies{
		Commands: []string{"nonexistent-tool-abc-123"},
	}

	err := checker.CheckAll(deps, map[string]interface{}{}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-tool-abc-123")
}

func TestDependencyChecker_CheckAll_VariablesFail(t *testing.T) {
	checker := NewDependencyChecker()

	deps := &core.Dependencies{
		Commands: []string{"echo"},
		Variables: []core.VariableDep{
			{
				Name:     "required_var",
				Type:     core.VarTypeString,
				Required: true,
			},
		},
	}

	// Missing required variable
	err := checker.CheckAll(deps, map[string]interface{}{}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required_var")
}

// DependencyError Tests

func TestDependencyError_Error_Missing(t *testing.T) {
	err := &DependencyError{
		Type:    "commands",
		Missing: []string{"cmd1", "cmd2"},
	}

	assert.Contains(t, err.Error(), "missing")
	assert.Contains(t, err.Error(), "commands")
	assert.Contains(t, err.Error(), "cmd1")
	assert.Contains(t, err.Error(), "cmd2")
}

func TestDependencyError_Error_Validation(t *testing.T) {
	err := &DependencyError{
		Type:   "variables",
		Errors: []string{"var1 is invalid", "var2 is missing"},
	}

	assert.Contains(t, err.Error(), "validation errors")
	assert.Contains(t, err.Error(), "var1 is invalid")
}

// Helper Function Tests

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected bool
	}{
		{"example.com", true},
		{"sub.example.com", true},
		{"a.b.c.example.com", true},
		{"example.co.uk", true},
		{"123.example.com", true},
		{"", false},
		{"invalid", false},
		{"-invalid.com", false},
		{"invalid-.com", false},
		{"example", false},
	}

	for _, tc := range tests {
		t.Run(tc.domain, func(t *testing.T) {
			result, err := core.MatchesVariableType(tc.domain, core.VarTypeDomain)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result, "domain: %s", tc.domain)
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"123", true},
		{"-123", true},
		{"12.34", true},
		{"-12.34", true},
		{"0", true},
		{"", false},
		{"abc", false},
		{"12abc", false},
		{"12.34.56", false},
	}

	for _, tc := range tests {
		t.Run(tc.value, func(t *testing.T) {
			result, err := core.MatchesVariableType(tc.value, core.VarTypeNumber)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result, "value: %s", tc.value)
		})
	}
}

// Global Function Tests

func TestCheckDependencies(t *testing.T) {
	deps := &core.Dependencies{
		Commands: []string{"echo"},
	}

	err := CheckDependencies(deps, map[string]interface{}{}, "")
	assert.NoError(t, err)
}

func TestCheckDependencies_Nil(t *testing.T) {
	err := CheckDependencies(nil, map[string]interface{}{}, "")
	assert.NoError(t, err)
}
