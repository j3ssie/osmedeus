package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/j3ssie/osmedeus/v5/internal/installer"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/j3ssie/osmedeus/v5/public"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// embedSkillsRoot is the directory inside public.EmbedFS holding the
// coding-agent skill bundles.
const embedSkillsRoot = "skills"

// defaultInstallSkill is installed when `skills install` runs without an
// explicit skill name.
const defaultInstallSkill = "osmedeus-expert"

// Skill directories understood by the supported coding agents. ".agents" is
// the cross-agent convention; ".claude" is Claude Code specific.
var (
	claudeSkillsSubdir = filepath.Join(".claude", "skills")
	agentsSkillsSubdir = filepath.Join(".agents", "skills")
)

var (
	skillsFull  bool
	skillsAll   bool
	skillsAgent string
	skillsScope string
	skillsDir   string
)

// bundledSkill is a parsed skill bundle shipped inside the binary.
type bundledSkill struct {
	Name        string   // directory name, e.g. "osmedeus-expert"
	Description string   // from SKILL.md frontmatter
	EmbedDir    string   // path inside public.EmbedFS, e.g. "skills/osmedeus-expert"
	References  []string // reference paths relative to EmbedDir, e.g. "references/cli-flags.md"
}

var skillsCmd = &cobra.Command{
	Use:     "skills",
	Aliases: []string{"skill"},
	Short:   "List, inspect, and install bundled coding-agent skills",
	Long:    UsageSkills(),
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSkillsList()
	},
}

var skillsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all bundled skills",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSkillsList()
	},
}

var skillsGetCmd = &cobra.Command{
	Use:   "get [name...]",
	Short: "Print a skill's full content to stdout",
	Long: `Print a skill's SKILL.md to stdout. Pass --full to also include its
reference files, or --all to output every bundled skill.

Useful for piping a skill into an agent that does not read skill directories:

  osmedeus skills get osmedeus-expert --full | your-agent`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSkillsGet(args)
	},
}

var skillsInstallCmd = &cobra.Command{
	Use:     "install [name...]",
	Aliases: []string{"add"},
	Short:   "Install skill bundle(s) into a coding agent's skills directory",
	Long: `Copy one or more bundled skills into a coding agent's skills directory so
the agent can auto-trigger on them.

With no name, installs the '` + defaultInstallSkill + `' skill. Pass --all to install
every bundle, or name specific skills.

Destination is chosen from --agent and --scope:

  --agent claude          .claude/skills/   (project)   ~/.claude/skills/   (global)
  --agent codex|agents    .agents/skills/   (project)   ~/.agents/skills/   (global)

An already-installed skill is skipped unless --force is given. --dir overrides
the computed destination entirely.`,
	Example: `  # Install the default skill into ./.claude/skills/
  osmedeus skills install

  # Install globally so every project sees it
  osmedeus skills install --scope global

  # Install for a Codex-style agent reading .agents/skills/
  osmedeus skills install --agent codex

  # Install every bundled skill, overwriting existing copies
  osmedeus skills install --all --force

  # Install to an explicit directory
  osmedeus skills install --dir ~/my-agent/skills`,
	Args: cobra.ArbitraryArgs,
	RunE: RunSkillsInstall,
}

func init() {
	skillsCmd.AddCommand(skillsListCmd, skillsGetCmd, skillsInstallCmd)

	skillsGetCmd.Flags().BoolVar(&skillsFull, "full", false, "include reference files, not just SKILL.md")
	skillsGetCmd.Flags().BoolVar(&skillsAll, "all", false, "output every bundled skill")

	addSkillsInstallFlags(skillsInstallCmd)
}

// addSkillsInstallFlags registers the install flags on a command. Called for
// both `skills install` and its `install skills` alias, which share handlers
// and flag variables (only one can run per invocation).
func addSkillsInstallFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&skillsAgent, "agent", "claude", "target coding agent: claude, codex, or agents")
	cmd.Flags().StringVar(&skillsScope, "scope", "project", "install scope: project (current folder) or global (home dir)")
	cmd.Flags().BoolVar(&skillsAll, "all", false, "install every bundled skill")
	cmd.Flags().StringVar(&skillsDir, "dir", "", "override the destination directory (skips --agent/--scope resolution)")
	// Note: --force flag is global (defined in root.go)
}

// parseSkillDescription extracts the description from a SKILL.md's YAML
// frontmatter. Bundles are first-party and vendored in-tree, so malformed
// frontmatter is a sync bug to fix upstream, not to route around here — the
// bundle simply lists without a description.
func parseSkillDescription(raw []byte) string {
	content := strings.TrimSpace(strings.TrimPrefix(string(raw), "\ufeff"))
	if !strings.HasPrefix(content, "---") {
		return ""
	}
	// Take everything between the opening and closing delimiters.
	frontmatter, _, ok := strings.Cut(strings.TrimPrefix(content, "---"), "\n---")
	if !ok {
		return ""
	}

	var fm struct {
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return ""
	}
	return strings.TrimSpace(fm.Description)
}

// loadBundledSkills discovers and parses every skill bundle embedded under
// public.EmbedFS/skills. A bundle is any directory containing a SKILL.md, so
// adding a bundle upstream needs no code change here.
func loadBundledSkills() ([]bundledSkill, error) {
	entries, err := fs.ReadDir(public.EmbedFS, embedSkillsRoot)
	if err != nil {
		return nil, fmt.Errorf("read embedded skills: %w", err)
	}

	var out []bundledSkill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		embedDir := embedSkillsRoot + "/" + e.Name()
		raw, readErr := public.EmbedFS.ReadFile(embedDir + "/SKILL.md")
		if readErr != nil {
			continue // not a skill bundle
		}

		out = append(out, bundledSkill{
			Name:        e.Name(),
			Description: parseSkillDescription(raw),
			EmbedDir:    embedDir,
			References:  listBundleReferences(embedDir),
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// listBundleReferences returns the files under <embedDir>/references as paths
// relative to embedDir, sorted.
func listBundleReferences(embedDir string) []string {
	entries, err := fs.ReadDir(public.EmbedFS, embedDir+"/references")
	if err != nil {
		return nil
	}
	var refs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		refs = append(refs, "references/"+e.Name())
	}
	sort.Strings(refs)
	return refs
}

// findBundle returns the bundle with the given name (case-insensitive).
func findBundle(skills []bundledSkill, name string) (bundledSkill, bool) {
	for _, s := range skills {
		if strings.EqualFold(s.Name, name) {
			return s, true
		}
	}
	return bundledSkill{}, false
}

// bundleNames joins bundle names for error messages.
func bundleNames(skills []bundledSkill) string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return strings.Join(names, ", ")
}

// resolveNamedTargets maps explicit skill names to bundles, erroring on the
// first unknown name.
func resolveNamedTargets(skills []bundledSkill, names []string) ([]bundledSkill, error) {
	targets := make([]bundledSkill, 0, len(names))
	for _, n := range names {
		b, ok := findBundle(skills, n)
		if !ok {
			return nil, fmt.Errorf("unknown skill %q (available: %s)", n, bundleNames(skills))
		}
		targets = append(targets, b)
	}
	return targets, nil
}

// resolveTargets picks the bundles a subcommand should act on, applying the
// same precedence for get and install: --all wins, then explicit names, then
// fallbackName. An empty fallbackName means a name is required.
func resolveTargets(skills []bundledSkill, names []string, fallbackName string) ([]bundledSkill, error) {
	switch {
	case skillsAll:
		return skills, nil
	case len(names) > 0:
		return resolveNamedTargets(skills, names)
	case fallbackName == "":
		return nil, fmt.Errorf("provide a skill name or --all (available: %s)", bundleNames(skills))
	}

	b, ok := findBundle(skills, fallbackName)
	if !ok {
		return nil, fmt.Errorf("default skill %q is not bundled (available: %s)", fallbackName, bundleNames(skills))
	}
	return []bundledSkill{b}, nil
}

func runSkillsList() error {
	skills, err := loadBundledSkills()
	if err != nil {
		return err
	}

	if globalJSON {
		type jsonEntry struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			EmbedPath   string   `json:"embed_path"`
			References  []string `json:"references"`
		}
		entries := make([]jsonEntry, len(skills))
		for i, s := range skills {
			entries[i] = jsonEntry{s.Name, s.Description, s.EmbedDir, s.References}
		}
		return outputJSON(map[string]interface{}{"skills": entries, "total": len(entries)})
	}

	printer := terminal.NewPrinter()
	if len(skills) == 0 {
		printer.Info("No skills to show.")
		return nil
	}

	printer.Newline()
	printer.Info("%s bundled skill(s)", terminal.BoldCyan(fmt.Sprintf("%d", len(skills))))
	printer.Newline()

	// Fixed columns for name and refs; description absorbs the remaining width.
	const nameWidth, refsWidth = 24, 4
	descWidth := skillsDescriptionWidth(nameWidth + refsWidth)

	fmt.Printf("  %s  %s  %s\n",
		terminal.Bold(padRight("NAME", nameWidth)),
		terminal.Bold(padRight("DESCRIPTION", descWidth)),
		terminal.Bold("REFS"))

	for _, s := range skills {
		fmt.Printf("  %s  %s  %s\n",
			colorPadRight(s.Name, nameWidth, terminal.Cyan),
			truncatePad(s.Description, descWidth),
			terminal.Gray(fmt.Sprintf("%d", len(s.References))))
	}

	binaryPath := filepath.Base(os.Args[0])
	printer.Newline()
	printer.Info("Read a skill:    %s", terminal.Gray(binaryPath+" skills get "+skills[0].Name+" --full"))
	printer.Info("Install a skill: %s", terminal.Gray(binaryPath+" skills install --agent claude --scope project"))
	return nil
}

// skillsDescriptionWidth computes the description column width from the
// terminal size (or --width), leaving room for the other columns and gutters.
func skillsDescriptionWidth(otherColumns int) int {
	const minWidth, maxWidth = 30, 110

	effective := globalWidth
	if effective == 0 {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			effective = w
		}
	}
	if effective == 0 {
		return 60
	}

	// Gutters: 2 leading spaces plus 2 separators of 2 spaces each.
	return min(max(effective-otherColumns-6, minWidth), maxWidth)
}

func runSkillsGet(names []string) error {
	skills, err := loadBundledSkills()
	if err != nil {
		return err
	}

	targets, err := resolveTargets(skills, names, "")
	if err != nil {
		return fmt.Errorf("skills get: %w", err)
	}

	for i, b := range targets {
		if len(targets) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("===== %s =====\n\n", b.Name)
		}
		body, readErr := public.EmbedFS.ReadFile(b.EmbedDir + "/SKILL.md")
		if readErr != nil {
			return fmt.Errorf("read %s: %w", b.Name, readErr)
		}
		writeSkillBody(body)

		if !skillsFull {
			continue
		}
		for _, ref := range b.References {
			refBody, refErr := public.EmbedFS.ReadFile(b.EmbedDir + "/" + ref)
			if refErr != nil {
				continue
			}
			fmt.Printf("\n\n----- %s/%s -----\n\n", b.Name, ref)
			writeSkillBody(refBody)
		}
	}
	return nil
}

// writeSkillBody writes raw bytes to stdout, appending a newline only when the
// content does not already end with one.
func writeSkillBody(data []byte) {
	_, _ = os.Stdout.Write(data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		fmt.Println()
	}
}

// skillsInstallBaseDir resolves the parent directory that skill bundles are
// installed into, from --agent and --scope. Returns an absolute path.
func skillsInstallBaseDir(agent, scope string) (string, error) {
	var sub string
	switch strings.ToLower(agent) {
	case "claude":
		sub = claudeSkillsSubdir
	case "codex", "agents":
		sub = agentsSkillsSubdir
	default:
		return "", fmt.Errorf("unknown --agent %q (want: claude, codex, or agents)", agent)
	}

	switch strings.ToLower(scope) {
	case "project":
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve working directory: %w", err)
		}
		return filepath.Join(cwd, sub), nil
	case "global":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, sub), nil
	default:
		return "", fmt.Errorf("unknown --scope %q (want: project or global)", scope)
	}
}

// RunSkillsInstall installs bundled skills into a coding agent's skills
// directory. Exported so the `install skills` alias can share this handler.
func RunSkillsInstall(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	skills, err := loadBundledSkills()
	if err != nil {
		return err
	}

	targets, err := resolveTargets(skills, args, defaultInstallSkill)
	if err != nil {
		return err
	}

	// Resolve the destination base directory.
	baseDir := skillsDir
	if baseDir == "" {
		if baseDir, err = skillsInstallBaseDir(skillsAgent, skillsScope); err != nil {
			return err
		}
	} else if abs, absErr := filepath.Abs(installer.ExpandPath(baseDir)); absErr == nil {
		baseDir = abs
	}

	installed, skipped := 0, 0
	for _, b := range targets {
		dest := filepath.Join(baseDir, b.Name)
		if _, statErr := os.Stat(filepath.Join(dest, "SKILL.md")); statErr == nil && !globalForce {
			printer.Warning("%s already installed at %s (use --force to overwrite)",
				terminal.Cyan(b.Name), terminal.Gray(dest))
			skipped++
			continue
		}
		if copyErr := copyEmbeddedTree(b.EmbedDir, dest); copyErr != nil {
			return fmt.Errorf("install %s: %w", b.Name, copyErr)
		}
		printer.Success("Installed %s → %s", terminal.Cyan(b.Name), terminal.Gray(dest))
		installed++
	}

	fmt.Println()
	printer.Info("%d installed, %d skipped", installed, skipped)
	if installed > 0 {
		printer.Info("Your agent will auto-trigger on these skills when you ask about osmedeus")
	}
	return nil
}
