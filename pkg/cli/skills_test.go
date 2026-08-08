package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/j3ssie/osmedeus/v5/public"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetSkillsFlags restores the package-level flag vars that the skills
// commands bind to, so tests do not leak state into each other.
func resetSkillsFlags(t *testing.T) {
	t.Helper()
	reset := func() {
		skillsAll, skillsAgent, skillsScope, skillsDir, globalForce = false, "claude", "project", "", false
	}
	reset()
	t.Cleanup(reset)
}

func TestLoadBundledSkills(t *testing.T) {
	skills, err := loadBundledSkills()
	require.NoError(t, err)
	require.NotEmpty(t, skills, "expected at least one embedded skill bundle")

	b, ok := findBundle(skills, defaultInstallSkill)
	require.True(t, ok, "default install skill %q must be bundled", defaultInstallSkill)

	assert.Equal(t, "skills/"+defaultInstallSkill, b.EmbedDir)
	assert.NotEmpty(t, b.Description, "description should be parsed from SKILL.md frontmatter")
	assert.NotEmpty(t, b.References, "osmedeus-expert ships reference files")
	for _, ref := range b.References {
		assert.True(t, filepath.Ext(ref) == ".md", "unexpected reference file: %s", ref)
	}
}

func TestFindBundleIsCaseInsensitive(t *testing.T) {
	skills := []bundledSkill{{Name: "osmedeus-expert"}}

	_, ok := findBundle(skills, "OSMEDEUS-EXPERT")
	assert.True(t, ok)

	_, ok = findBundle(skills, "nope")
	assert.False(t, ok)
}

func TestParseSkillDescription(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "standard frontmatter",
			raw:  "---\nname: demo\ndescription: A demo skill\n---\n\n# Body\n",
			want: "A demo skill",
		},
		{
			name: "quoted description with colons",
			raw:  "---\nname: demo\ndescription: \"Use when: (1) a thing, (2) another\"\n---\nbody",
			want: "Use when: (1) a thing, (2) another",
		},
		{
			name: "leading BOM is tolerated",
			raw:  "\ufeff---\nname: demo\ndescription: bom skill\n---\nbody",
			want: "bom skill",
		},
		{
			name: "no frontmatter",
			raw:  "# Just a heading\n",
			want: "",
		},
		{
			name: "unterminated frontmatter",
			raw:  "---\nname: demo\ndescription: never closed\n",
			want: "",
		},
		{
			name: "malformed yaml degrades to empty",
			raw:  "---\n\tname: [unclosed\n---\nbody",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseSkillDescription([]byte(tt.raw)))
		})
	}
}

func TestResolveNamedTargetsUnknownName(t *testing.T) {
	skills := []bundledSkill{{Name: "osmedeus-expert"}}

	_, err := resolveNamedTargets(skills, []string{"osmedeus-expert", "ghost"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown skill "ghost"`)
	assert.Contains(t, err.Error(), "available: osmedeus-expert")
}

func TestSkillsInstallBaseDir(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		agent, scope string
		want         string
		wantErr      string
	}{
		{agent: "claude", scope: "project", want: filepath.Join(cwd, ".claude", "skills")},
		{agent: "claude", scope: "global", want: filepath.Join(home, ".claude", "skills")},
		{agent: "codex", scope: "project", want: filepath.Join(cwd, ".agents", "skills")},
		{agent: "agents", scope: "global", want: filepath.Join(home, ".agents", "skills")},
		{agent: "CLAUDE", scope: "PROJECT", want: filepath.Join(cwd, ".claude", "skills")},
		{agent: "vim", scope: "project", wantErr: `unknown --agent "vim"`},
		{agent: "claude", scope: "everywhere", wantErr: `unknown --scope "everywhere"`},
	}

	for _, tt := range tests {
		t.Run(tt.agent+"/"+tt.scope, func(t *testing.T) {
			got, dirErr := skillsInstallBaseDir(tt.agent, tt.scope)
			if tt.wantErr != "" {
				require.Error(t, dirErr)
				assert.Contains(t, dirErr.Error(), tt.wantErr)
				return
			}
			require.NoError(t, dirErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRunSkillsInstallToExplicitDir(t *testing.T) {
	resetSkillsFlags(t)
	dest := t.TempDir()
	skillsDir = dest

	require.NoError(t, RunSkillsInstall(skillsInstallCmd, nil))

	bundleDir := filepath.Join(dest, defaultInstallSkill)
	assert.FileExists(t, filepath.Join(bundleDir, "SKILL.md"))

	// References must be copied too, not just the top-level SKILL.md.
	refs, err := os.ReadDir(filepath.Join(bundleDir, "references"))
	require.NoError(t, err)
	assert.NotEmpty(t, refs)

	// Installed content must match what is embedded.
	onDisk, err := os.ReadFile(filepath.Join(bundleDir, "SKILL.md"))
	require.NoError(t, err)
	embedded, err := readEmbeddedSkillMD(defaultInstallSkill)
	require.NoError(t, err)
	assert.Equal(t, embedded, onDisk)
}

func TestRunSkillsInstallSkipsWithoutForce(t *testing.T) {
	resetSkillsFlags(t)
	dest := t.TempDir()
	skillsDir = dest

	require.NoError(t, RunSkillsInstall(skillsInstallCmd, nil))

	// Mark the installed copy so we can tell whether it was overwritten.
	installed := filepath.Join(dest, defaultInstallSkill, "SKILL.md")
	require.NoError(t, os.WriteFile(installed, []byte("stale"), 0644))

	// Without --force the existing bundle is left alone.
	require.NoError(t, RunSkillsInstall(skillsInstallCmd, nil))
	body, err := os.ReadFile(installed)
	require.NoError(t, err)
	assert.Equal(t, "stale", string(body), "expected install to skip an existing bundle")

	// With --force it is overwritten from the embedded copy.
	globalForce = true
	require.NoError(t, RunSkillsInstall(skillsInstallCmd, nil))
	body, err = os.ReadFile(installed)
	require.NoError(t, err)
	assert.NotEqual(t, "stale", string(body), "expected --force to overwrite")
}

func TestRunSkillsInstallUnknownName(t *testing.T) {
	resetSkillsFlags(t)
	skillsDir = t.TempDir()

	err := RunSkillsInstall(skillsInstallCmd, []string{"not-a-skill"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown skill "not-a-skill"`)
}

func TestRunSkillsInstallAll(t *testing.T) {
	resetSkillsFlags(t)
	dest := t.TempDir()
	skillsDir = dest
	skillsAll = true

	require.NoError(t, RunSkillsInstall(skillsInstallCmd, nil))

	skills, err := loadBundledSkills()
	require.NoError(t, err)
	for _, s := range skills {
		assert.FileExists(t, filepath.Join(dest, s.Name, "SKILL.md"))
	}
}

// readEmbeddedSkillMD is a test helper mirroring how install reads a bundle.
func readEmbeddedSkillMD(name string) ([]byte, error) {
	skills, err := loadBundledSkills()
	if err != nil {
		return nil, err
	}
	b, ok := findBundle(skills, name)
	if !ok {
		return nil, os.ErrNotExist
	}
	return public.EmbedFS.ReadFile(b.EmbedDir + "/SKILL.md")
}
