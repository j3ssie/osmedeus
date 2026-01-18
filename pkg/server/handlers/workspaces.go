package handlers

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

func listWorkspaceDirs(workspacesDir string) ([]string, error) {
	if workspacesDir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(workspacesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !isValidWorkspaceName(name) {
			continue
		}
		if entry.IsDir() {
			names = append(names, name)
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(filepath.Join(workspacesDir, name))
			if err == nil && info.IsDir() {
				names = append(names, name)
			}
		}
	}

	sort.Strings(names)
	return names, nil
}

func resolveWorkspacesDirForListing(cfg *config.Config) string {
	configured := ""
	if cfg != nil {
		configured = cfg.GetWorkspacesDir()
	}
	if configured != "" {
		if _, err := os.Stat(configured); err == nil {
			return configured
		}
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		candidate := filepath.Join(home, "workspaces-osmedeus")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	if _, err := os.Stat("/workspaces-osmedeus"); err == nil {
		return "/workspaces-osmedeus"
	}

	return configured
}

type filesystemWorkspaceRecord struct {
	Name        string   `json:"name"`
	LocalPath   string   `json:"local_path,omitempty"`
	DataSource  string   `json:"data_source"`
	TotalAssets int      `json:"total_assets"`
	Tags        []string `json:"tags"`
}

// ListWorkspaces handles listing all workspaces
// @Summary List all workspaces
// @Description Get a list of all run workspaces. By default returns full workspace records from database. Use filesystem=true to list workspaces derived from assets.
// @Tags Workspaces
// @Produce json
// @Param filesystem query bool false "List workspaces from filesystem/assets instead of workspaces table" default(false)
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return (max 10000)" default(20)
// @Success 200 {object} map[string]interface{} "List of workspaces"
// @Failure 500 {object} map[string]interface{} "Failed to read workspaces"
// @Security BearerAuth
// @Router /osm/api/workspaces [get]
func ListWorkspaces(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
		filesystem := c.Query("filesystem", "false") == "true"
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		limit, _ := strconv.Atoi(c.Query("limit", "20"))

		// Validate pagination
		if offset < 0 {
			offset = 0
		}
		if limit <= 0 {
			limit = 20
		}
		if limit > 10000 {
			limit = 10000
		}

		ctx := context.Background()

		// Return workspaces based on mode
		if filesystem {
			assetWorkspaces, err := database.ListAllWorkspacesFromAssets(ctx)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			inAssetsDBByName := make(map[string]bool, len(assetWorkspaces))
			for _, ws := range assetWorkspaces {
				inAssetsDBByName[ws.Name] = true
			}

			workspacesDir := resolveWorkspacesDirForListing(cfg)
			workspaceDirs, err := listWorkspaceDirs(workspacesDir)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			assetCountByName := make(map[string]int, len(assetWorkspaces))
			for _, ws := range assetWorkspaces {
				assetCountByName[ws.Name] = ws.AssetCount
			}

			hasDirByName := make(map[string]bool, len(workspaceDirs))
			for _, name := range workspaceDirs {
				hasDirByName[name] = true
				if _, ok := assetCountByName[name]; !ok {
					assetCountByName[name] = 0
				}
			}

			records := make([]filesystemWorkspaceRecord, 0, len(assetCountByName))
			for name, assetCount := range assetCountByName {
				tags := []string{"filesystem"}
				if hasDirByName[name] && !inAssetsDBByName[name] {
					tags = append(tags, "filesystem-only")
				}
				rec := filesystemWorkspaceRecord{
					Name:        name,
					DataSource:  "filesystem",
					TotalAssets: assetCount,
					Tags:        tags,
				}
				if hasDirByName[name] {
					rec.LocalPath = filepath.Join(workspacesDir, name)
				}
				records = append(records, rec)
			}
			sort.Slice(records, func(i, j int) bool {
				return records[i].Name < records[j].Name
			})

			totalCount := len(records)
			if offset > totalCount {
				offset = totalCount
			}
			end := offset + limit
			if end > totalCount {
				end = totalCount
			}
			page := records[offset:end]

			return c.JSON(fiber.Map{
				"data":           page,
				"workspaces_dir": workspacesDir,
				"pagination": fiber.Map{
					"total":  totalCount,
					"offset": offset,
					"limit":  limit,
				},
			})
		}

		// Default: Get full workspace records from workspaces table
		result, err := database.ListWorkspacesFullFromDB(ctx, offset, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data": result.Data,
			"pagination": fiber.Map{
				"total":  result.TotalCount,
				"offset": result.Offset,
				"limit":  result.Limit,
			},
		})
	}
}

// ListWorkspaceNames handles listing workspace names
// @Summary List workspace names
// @Description Get a sorted list of workspace names from the database
// @Tags Workspaces
// @Produce json
// @Success 200 {array} string "Workspace names"
// @Failure 500 {object} map[string]interface{} "Failed to list workspace names"
// @Security BearerAuth
// @Router /osm/api/workspace-names [get]
func ListWorkspaceNames(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		db := database.GetDB()
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Database not connected",
			})
		}

		var names []string
		if err := db.NewSelect().Model((*database.Workspace)(nil)).Column("name").Order("name ASC").Scan(ctx, &names); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(names)
	}
}

// isValidWorkspaceName validates workspace name to prevent path traversal
func isValidWorkspaceName(name string) bool {
	// Reject empty, ".", "..", or names containing path separators
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	if strings.Contains(name, "..") {
		return false
	}
	return true
}

// isPathUnderWorkspace ensures the file path is within the workspace folder
func isPathUnderWorkspace(filePath, workspacePath string) bool {
	if filePath == "" || workspacePath == "" {
		return false
	}
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	absWorkspace, err := filepath.Abs(workspacePath)
	if err != nil {
		return false
	}

	realFile, err := filepath.EvalSymlinks(absFile)
	if err == nil {
		absFile = realFile
	}
	realWorkspace, err := filepath.EvalSymlinks(absWorkspace)
	if err == nil {
		absWorkspace = realWorkspace
	}

	return strings.HasPrefix(absFile, absWorkspace+string(filepath.Separator)) || absFile == absWorkspace
}
