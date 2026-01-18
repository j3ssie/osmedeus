package handlers

import (
	"os/exec"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/installer"
	"github.com/j3ssie/osmedeus/v5/public"
)

// GetRegistryInfo returns binary registry with installation status
// @Summary Get registry info
// @Description Get binary registry with mode support (direct-fetch or nix-build)
// @Tags Install
// @Produce json
// @Param registry_mode query string false "Registry mode: direct-fetch or nix-build" default(direct-fetch)
// @Success 200 {object} map[string]interface{} "Registry data"
// @Failure 500 {object} map[string]interface{} "Failed to load registry"
// @Security BearerAuth
// @Router /osm/api/registry-info [get]
func GetRegistryInfo(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		registryMode := c.Query("registry_mode", "direct-fetch")

		switch registryMode {
		case "nix-build":
			return getNixBuildRegistry(c)
		case "direct-fetch":
			fallthrough
		default:
			return getDirectFetchRegistry(c)
		}
	}
}

// getDirectFetchRegistry returns the direct-fetch registry (existing behavior)
func getDirectFetchRegistry(c *fiber.Ctx) error {
	registry, err := installer.LoadRegistry("", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load registry: " + err.Error(),
		})
	}

	// Build response with installation status for each binary
	binariesWithStatus := make(map[string]BinaryStatusEntry)
	for name, entry := range registry {
		path, _ := exec.LookPath(name)
		binariesWithStatus[name] = BinaryStatusEntry{
			Desc:                entry.Desc,
			RepoLink:            entry.RepoLink,
			Version:             entry.Version,
			Tags:                entry.Tags,
			ValidateCommand:     entry.ValidateCommand,
			Linux:               entry.Linux,
			Darwin:              entry.Darwin,
			Windows:             entry.Windows,
			CommandLinux:        entry.CommandLinux,
			CommandDarwin:       entry.CommandDarwin,
			CommandDual:         entry.CommandDual,
			MultiCommandsLinux:  entry.MultiCommandsLinux,
			MultiCommandsDarwin: entry.MultiCommandsDarwin,
			Installed:           installer.IsBinaryInstalled(name, &entry),
			Path:                path,
		}
	}

	return c.JSON(fiber.Map{
		"registry_mode": "direct-fetch",
		"registry_url":  installer.DefaultRegistryURL,
		"binaries":      binariesWithStatus,
	})
}

// getNixBuildRegistry returns Nix flake binaries with registry metadata
func getNixBuildRegistry(c *fiber.Ctx) error {
	// Parse flake.nix
	flakeContent, err := public.GetFlakeNix()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to read flake.nix: " + err.Error(),
		})
	}

	categories, err := installer.ParseFlakeNixBinariesFromString(string(flakeContent))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to parse flake.nix: " + err.Error(),
		})
	}

	// Load registry for metadata (desc, tags)
	registry, _ := installer.LoadRegistry("", nil)

	// Build response with categories and tool metadata
	categoriesData := make([]map[string]interface{}, 0)
	for _, cat := range categories {
		toolsData := make([]map[string]interface{}, 0)
		for _, tool := range cat.Tools {
			// Get entry for validation command check
			var entryPtr *installer.BinaryEntry
			if registry != nil {
				if entry, ok := registry[tool]; ok {
					entryPtr = &entry
				}
			}

			toolData := map[string]interface{}{
				"name":      tool,
				"installed": installer.IsBinaryInstalled(tool, entryPtr),
			}
			if entryPtr != nil {
				toolData["desc"] = entryPtr.Desc
				toolData["tags"] = entryPtr.Tags
				toolData["version"] = entryPtr.Version
				toolData["repo_link"] = entryPtr.RepoLink
				if entryPtr.ValidateCommand != "" {
					toolData["valide-command"] = entryPtr.ValidateCommand
				}
			}
			if path, err := exec.LookPath(tool); err == nil {
				toolData["path"] = path
			}
			toolsData = append(toolsData, toolData)
		}
		catData := map[string]interface{}{
			"name":  cat.Name,
			"tools": toolsData,
		}
		categoriesData = append(categoriesData, catData)
	}

	return c.JSON(fiber.Map{
		"registry_mode": "nix-build",
		"nix_installed": installer.IsNixInstalled(),
		"categories":    categoriesData,
	})
}

// RegistryInstall handles binary or workflow installation via API
// @Summary Install binaries or workflows
// @Description Install binaries from registry or workflows from git/zip URL. Supports direct-fetch and nix-build modes.
// @Tags Install
// @Accept json
// @Produce json
// @Param request body InstallRequest true "Installation configuration"
// @Success 200 {object} map[string]interface{} "Installation result"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Installation failed"
// @Security BearerAuth
// @Router /osm/api/registry-install [post]
func RegistryInstall(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req InstallRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		// Default registry mode
		if req.RegistryMode == "" {
			req.RegistryMode = "direct-fetch"
		}

		// Validate type
		if req.Type != "binary" && req.Type != "workflow" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "type must be 'binary' or 'workflow'",
			})
		}

		inst := installer.NewInstaller(cfg.BaseFolder, cfg.WorkflowsPath, cfg.BinariesPath, nil)

		switch req.Type {
		case "binary":
			return installBinaries(c, cfg, inst, req)
		case "workflow":
			return installWorkflow(c, inst, req)
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid installation type",
			})
		}
	}
}

// installBinaries handles binary installation with mode support
func installBinaries(c *fiber.Ctx, cfg *config.Config, inst *installer.Installer, req InstallRequest) error {
	switch req.RegistryMode {
	case "nix-build":
		return installBinariesViaNix(c, cfg, inst, req)
	case "direct-fetch":
		fallthrough
	default:
		return installBinariesDirectFetch(c, inst, req)
	}
}

// installBinariesDirectFetch handles binary installation via direct download
func installBinariesDirectFetch(c *fiber.Ctx, inst *installer.Installer, req InstallRequest) error {
	registryURL := req.RegistryURL
	if registryURL == "" {
		registryURL = installer.DefaultRegistryURL
	}

	// Load registry first
	registry, err := installer.LoadRegistry(registryURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load registry: " + err.Error(),
		})
	}

	var installed []string
	var failed []map[string]string

	if req.InstallAll {
		// Install all binaries from registry
		for name := range registry {
			if err := installer.InstallBinary(name, registry, inst.BinariesFolder, nil); err != nil {
				failed = append(failed, map[string]string{
					"name":  name,
					"error": err.Error(),
				})
			} else {
				installed = append(installed, name)
			}
		}
	} else {
		// Install specified binaries
		if len(req.Names) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "names array is required for binary installation (or set install_all=true)",
			})
		}

		for _, name := range req.Names {
			if err := installer.InstallBinary(name, registry, inst.BinariesFolder, nil); err != nil {
				failed = append(failed, map[string]string{
					"name":  name,
					"error": err.Error(),
				})
			} else {
				installed = append(installed, name)
			}
		}
	}

	response := fiber.Map{
		"message":         "Binary installation completed",
		"registry_mode":   "direct-fetch",
		"installed":       installed,
		"installed_count": len(installed),
		"binaries_folder": inst.BinariesFolder,
	}

	if len(failed) > 0 {
		response["failed"] = failed
		response["failed_count"] = len(failed)
	}

	return c.JSON(response)
}

// installBinariesViaNix handles binary installation via Nix
func installBinariesViaNix(c *fiber.Ctx, cfg *config.Config, _ *installer.Installer, req InstallRequest) error {
	// Check if Nix is installed
	if !installer.IsNixInstalled() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Nix is not installed. Install Nix first or use registry_mode=direct-fetch",
		})
	}

	// Get binaries folder
	binariesFolder := cfg.BinariesPath
	if binariesFolder == "" {
		binariesFolder = filepath.Join(cfg.BaseFolder, "binaries")
	}

	var names []string
	if req.InstallAll {
		// Get all binaries from flake
		flakeContent, err := public.GetFlakeNix()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to read flake.nix: " + err.Error(),
			})
		}
		categories, err := installer.ParseFlakeNixBinariesFromString(string(flakeContent))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to parse flake.nix: " + err.Error(),
			})
		}
		names = installer.GetAllFlakeBinaries(categories)
	} else {
		if len(req.Names) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "names array is required for binary installation (or set install_all=true)",
			})
		}
		names = req.Names
	}

	// Install each binary via Nix
	var installed []string
	var failed []map[string]string

	for _, name := range names {
		if err := installer.InstallBinaryViaNix(name, "", binariesFolder); err != nil {
			failed = append(failed, map[string]string{
				"name":  name,
				"error": err.Error(),
			})
		} else {
			installed = append(installed, name)
		}
	}

	response := fiber.Map{
		"message":         "Nix binary installation completed",
		"registry_mode":   "nix-build",
		"installed":       installed,
		"installed_count": len(installed),
		"binaries_folder": binariesFolder,
	}

	if len(failed) > 0 {
		response["failed"] = failed
		response["failed_count"] = len(failed)
	}

	return c.JSON(response)
}

// installWorkflow handles workflow installation
func installWorkflow(c *fiber.Ctx, inst *installer.Installer, req InstallRequest) error {
	if req.Source == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "source is required for workflow installation (git URL, zip URL, or local path)",
		})
	}

	if err := inst.InstallWorkflow(req.Source); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to install workflow: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":         "Workflow installed successfully",
		"source":          req.Source,
		"workflow_folder": inst.WorkflowFolder,
	})
}
