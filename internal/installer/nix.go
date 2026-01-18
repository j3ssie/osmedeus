package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/orivej/go-nix/nix/parser"
	"go.uber.org/zap"
)

// NixInstallOutput stores the last Nix installation output for display
var NixInstallOutput string

// DeterminateNixInstallerURL is the URL for the Determinate Systems Nix installer
const DeterminateNixInstallerURL = "https://install.determinate.systems/nix"

const NixInstallCommand = "curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install linux --extra-conf \"sandbox = false\" --init none --no-confirm"

const NixInstallCommandPretty = `curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install linux \
    --extra-conf "sandbox = false" \
    --init none \
    --no-confirm`

const nixInstallCommandDefault = "curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install --no-confirm"

const nixInstallCommandDefaultPretty = `curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install \
    --no-confirm`

func NixInstallCommandPrettyForHost() string {
	if runtime.GOOS == "linux" {
		return NixInstallCommandPretty
	}
	return nixInstallCommandDefaultPretty
}

// IsNixInstalled checks if Nix is available in PATH
func IsNixInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// InstallNix installs Nix using the Determinate Systems installer
func InstallNix() error {
	if IsNixInstalled() {
		logger.Get().Info("Nix is already installed")
		return nil
	}

	logger.Get().Info("Installing Nix via Determinate Systems installer...")

	installCmd := nixInstallCommandDefault
	if runtime.GOOS == "linux" {
		installCmd = NixInstallCommand
	}

	cmd := exec.Command("sh", "-c", installCmd)
	if os.Getenv("OSMEDEUS_SILENT") == "1" {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Nix: %w", err)
	}

	logger.Get().Info("Nix installed successfully")
	return nil
}

// InstallBinaryViaNix installs a binary using `nix profile add` and copies it to binariesFolder
// Returns the captured output for display purposes
func InstallBinaryViaNix(binaryName string, nixPackage string, binariesFolder string) error {
	if !IsNixInstalled() {
		return fmt.Errorf("nix is not installed")
	}

	// Use nix_package from registry if specified, otherwise use binary name
	pkgName := nixPackage
	if pkgName == "" {
		pkgName = binaryName
	}

	logger.Get().Info("Installing via Nix",
		zap.String("binary", binaryName),
		zap.String("nix_package", pkgName))

	installable := pkgName
	if !strings.Contains(installable, "#") {
		installable = fmt.Sprintf("nixpkgs#%s", pkgName)
	}

	cmd := exec.Command("nix", "profile", "add", installable)

	// Capture output for gray display
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Print output even on failure for debugging
		if len(output) > 0 {
			fmt.Print(string(output))
		}
		return fmt.Errorf("failed to add %s via Nix: %w", pkgName, err)
	}

	// Store output for caller to display
	NixInstallOutput = string(output)

	// Copy binary from Nix profile to binaries folder
	if binariesFolder != "" {
		if err := copyNixBinaryToFolder(binaryName, binariesFolder); err != nil {
			return fmt.Errorf("failed to copy binary to folder: %w", err)
		}
	}

	return nil
}

// copyNixBinaryToFolder finds a binary installed by Nix and copies it to the target folder
func copyNixBinaryToFolder(binaryName string, binariesFolder string) error {
	// Find the binary path using 'which'
	cmd := exec.Command("which", binaryName)
	output, err := cmd.Output()
	if err != nil {
		// Try common Nix profile paths
		homeDir, _ := os.UserHomeDir()
		nixPaths := []string{
			filepath.Join(homeDir, ".nix-profile", "bin", binaryName),
			filepath.Join("/nix/var/nix/profiles/default/bin", binaryName),
		}
		for _, p := range nixPaths {
			if _, statErr := os.Stat(p); statErr == nil {
				output = []byte(p)
				err = nil
				break
			}
		}
		if err != nil {
			return fmt.Errorf("binary %s not found after Nix install", binaryName)
		}
	}

	// Remove trailing newline and clean path
	srcPath := strings.TrimSpace(string(output))

	// Ensure binaries folder exists
	if err := os.MkdirAll(binariesFolder, 0755); err != nil {
		return fmt.Errorf("failed to create binaries folder: %w", err)
	}

	destPath := filepath.Join(binariesFolder, binaryName)

	logger.Get().Info("Copying binary to external folder",
		zap.String("src", srcPath),
		zap.String("dest", destPath))

	// Copy the file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination binary: %w", err)
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	return nil
}

// GetNixPackageName returns the Nix package name for a binary
// Returns the nix_package field if set, otherwise returns the binary name
func GetNixPackageName(entry BinaryEntry, binaryName string) string {
	if entry.NixPackage != "" {
		return entry.NixPackage
	}
	return binaryName
}

// FlakeToolCategory represents a category of tools in a Nix flake
type FlakeToolCategory struct {
	Name  string   // Category name (derived from variable name)
	Tools []string // List of package names
}

// ParseFlakeNixBinaries parses a flake.nix file and extracts tool categories
func ParseFlakeNixBinaries(flakePath string) ([]FlakeToolCategory, error) {
	p, err := parser.ParseFile(flakePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flake.nix: %w", err)
	}

	if p.Result == nil {
		return nil, fmt.Errorf("empty parse result")
	}

	var categories []FlakeToolCategory
	extractToolCategories(p, p.Result, &categories)

	return categories, nil
}

// ParseFlakeNixBinariesFromString parses flake.nix content from a string
func ParseFlakeNixBinariesFromString(content string) ([]FlakeToolCategory, error) {
	p, err := parser.ParseString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flake.nix: %w", err)
	}

	if p.Result == nil {
		return nil, fmt.Errorf("empty parse result")
	}

	var categories []FlakeToolCategory
	extractToolCategories(p, p.Result, &categories)

	return categories, nil
}

// extractToolCategories traverses the AST to find tool category bindings
func extractToolCategories(p *parser.Parser, node *parser.Node, categories *[]FlakeToolCategory) {
	if node == nil {
		return
	}

	// Look for BindNode which represents `name = value;`
	if node.Type == parser.BindNode && len(node.Nodes) >= 2 {
		// First child is the attribute path (name), second is the value
		nameNode := node.Nodes[0]
		valueNode := node.Nodes[1]

		// Get the binding name
		bindName := extractBindingName(p, nameNode)

		// Check if this looks like a tools list (ends with "Tools")
		if strings.HasSuffix(bindName, "Tools") {
			tools := extractToolsFromValue(p, valueNode)
			if len(tools) > 0 {
				// Convert camelCase to readable name
				categoryName := formatCategoryName(bindName)
				*categories = append(*categories, FlakeToolCategory{
					Name:  categoryName,
					Tools: tools,
				})
			}
		}
	}

	// Recursively process child nodes
	for _, child := range node.Nodes {
		extractToolCategories(p, child, categories)
	}
}

// extractBindingName extracts the name from an AttrPathNode or IDNode
func extractBindingName(p *parser.Parser, node *parser.Node) string {
	if node == nil {
		return ""
	}

	if node.Type == parser.IDNode && len(node.Tokens) > 0 {
		return p.TokenString(node.Tokens[0])
	}

	if node.Type == parser.AttrPathNode && len(node.Nodes) > 0 {
		return extractBindingName(p, node.Nodes[0])
	}

	return ""
}

// extractToolsFromValue extracts package identifiers from a list or with expression
func extractToolsFromValue(p *parser.Parser, node *parser.Node) []string {
	if node == nil {
		return nil
	}

	var tools []string

	switch node.Type {
	case parser.ListNode:
		// Direct list: [ amass subfinder ... ]
		for _, child := range node.Nodes {
			if child.Type == parser.IDNode && len(child.Tokens) > 0 {
				tools = append(tools, p.TokenString(child.Tokens[0]))
			}
		}
	case parser.WithNode:
		// with pkgs; [ ... ] - the second child is the list
		if len(node.Nodes) >= 2 {
			return extractToolsFromValue(p, node.Nodes[1])
		}
	case parser.ParensNode:
		// Parenthesized expression
		if len(node.Nodes) > 0 {
			return extractToolsFromValue(p, node.Nodes[0])
		}
	}

	return tools
}

// formatCategoryName converts "subdomainTools" to "Subdomain"
func formatCategoryName(varName string) string {
	// Remove "Tools" suffix
	name := strings.TrimSuffix(varName, "Tools")
	if name == "" {
		return "Tools"
	}

	// Split camelCase into words
	var words []string
	current := ""
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			if current != "" {
				words = append(words, current)
			}
			current = string(r)
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}

	// Capitalize first letter of each word
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}

	return strings.Join(words, " ")
}

// GetAllFlakeBinaries returns all binary names from a flake, sorted alphabetically
func GetAllFlakeBinaries(categories []FlakeToolCategory) []string {
	seen := make(map[string]bool)
	var binaries []string

	for _, cat := range categories {
		for _, tool := range cat.Tools {
			if !seen[tool] {
				seen[tool] = true
				binaries = append(binaries, tool)
			}
		}
	}

	sort.Strings(binaries)
	return binaries
}
