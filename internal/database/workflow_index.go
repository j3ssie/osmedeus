package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/uptrace/bun"
)

// WorkflowQuery holds query parameters for listing workflows from DB
type WorkflowQuery struct {
	Tags   []string // Filter by tags (any match)
	Kind   string   // Filter by kind (module/flow)
	Search string   // Search in name/description
	Offset int
	Limit  int
}

// WorkflowMetaResult holds paginated workflow metadata results
type WorkflowMetaResult struct {
	Data       []WorkflowMeta `json:"data"`
	TotalCount int            `json:"total_count"`
	Offset     int            `json:"offset"`
	Limit      int            `json:"limit"`
}

// IndexResult holds the result of a workflow indexing operation
type IndexResult struct {
	Added   int      `json:"added"`
	Updated int      `json:"updated"`
	Removed int      `json:"removed"`
	Errors  []string `json:"errors,omitempty"`
}

// IndexWorkflowsFromFilesystem scans workflow directory and updates database
func IndexWorkflowsFromFilesystem(ctx context.Context, workflowsPath string, force bool) (*IndexResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &IndexResult{}

	// Load all workflows from filesystem
	loader := parser.NewLoader(workflowsPath)
	workflows, err := loader.LoadAllWorkflows()
	if err != nil {
		return nil, fmt.Errorf("failed to load workflows: %w", err)
	}

	// Track which workflows we've seen
	seenNames := make(map[string]bool)

	// Process each workflow
	for _, w := range workflows {
		seenNames[w.Name] = true

		// Check if workflow already exists
		var existing WorkflowMeta
		existsErr := db.NewSelect().Model(&existing).Where("name = ?", w.Name).Scan(ctx)
		existed := existsErr == nil

		if err := upsertWorkflowMeta(ctx, w, force); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", w.Name, err))
		} else {
			if existed {
				result.Updated++
			} else {
				result.Added++
			}
		}
	}

	// Remove workflows that no longer exist on filesystem
	var allMeta []WorkflowMeta
	if err := db.NewSelect().Model(&allMeta).Scan(ctx); err == nil {
		for _, meta := range allMeta {
			if !seenNames[meta.Name] {
				_, err := db.NewDelete().Model(&meta).Where("id = ?", meta.ID).Exec(ctx)
				if err == nil {
					result.Removed++
				}
			}
		}
	}

	return result, nil
}

// upsertWorkflowMeta inserts or updates a workflow metadata record
func upsertWorkflowMeta(ctx context.Context, w *core.Workflow, force bool) error {
	// Check if workflow already exists
	var existing WorkflowMeta
	err := db.NewSelect().Model(&existing).Where("name = ?", w.Name).Scan(ctx)

	// If exists and checksum unchanged (unless force), skip
	if err == nil && !force && existing.Checksum == w.Checksum {
		return nil
	}

	// Serialize params to JSON
	paramsJSON := ""
	if w.Params != nil {
		if data, err := json.Marshal(w.Params); err == nil {
			paramsJSON = string(data)
		}
	}

	now := time.Now()

	if err == nil {
		// Update existing
		existing.Kind = string(w.Kind)
		existing.Description = w.Description
		existing.FilePath = w.FilePath
		existing.Checksum = w.Checksum
		existing.Tags = w.Tags
		existing.Hidden = w.Hidden
		existing.StepCount = len(w.Steps)
		existing.ModuleCount = len(w.Modules)
		existing.HookCount = w.HookCount()
		existing.ParamsJSON = paramsJSON
		existing.IndexedAt = now
		existing.UpdatedAt = now

		_, err = db.NewUpdate().Model(&existing).WherePK().Exec(ctx)
		if err == nil {
			// Invalidate cache after successful update
			if cache := GetCache(); cache != nil {
				cache.InvalidateWorkflowMeta(w.Name)
			}
		}
		return err
	}

	// Insert new
	meta := &WorkflowMeta{
		Name:        w.Name,
		Kind:        string(w.Kind),
		Description: w.Description,
		FilePath:    w.FilePath,
		Checksum:    w.Checksum,
		Tags:        w.Tags,
		Hidden:      w.Hidden,
		StepCount:   len(w.Steps),
		ModuleCount: len(w.Modules),
		HookCount:   w.HookCount(),
		ParamsJSON:  paramsJSON,
		IndexedAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = db.NewInsert().Model(meta).Exec(ctx)
	if err == nil {
		// Invalidate cache after successful insert (in case of stale negative cache)
		if cache := GetCache(); cache != nil {
			cache.InvalidateWorkflowMeta(w.Name)
		}
	}
	return err
}

// ListWorkflowsFromDB returns paginated workflow metadata from database
func ListWorkflowsFromDB(ctx context.Context, query WorkflowQuery) (*WorkflowMetaResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &WorkflowMetaResult{
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	if result.Limit <= 0 {
		result.Limit = 20
	}
	if result.Limit > 10000 {
		result.Limit = 10000
	}

	// Build base query
	baseQuery := db.NewSelect().Model((*WorkflowMeta)(nil))

	// Filter out hidden workflows by default
	baseQuery = baseQuery.Where("hidden = ? OR hidden IS NULL", false)

	// Apply filters
	if query.Kind != "" {
		baseQuery = baseQuery.Where("kind = ?", query.Kind)
	}

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		baseQuery = baseQuery.Where("(name LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
	}

	// Tag filtering - check if any tag matches
	if len(query.Tags) > 0 {
		baseQuery = baseQuery.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for _, tag := range query.Tags {
				// For SQLite JSON, use json_each to search array
				if IsSQLite() {
					q = q.WhereOr("EXISTS (SELECT 1 FROM json_each(tags) WHERE value = ?)", tag)
				} else {
					// For PostgreSQL, use @> operator
					q = q.WhereOr("tags @> ?", fmt.Sprintf(`["%s"]`, tag))
				}
			}
			return q
		})
	}

	// Get total count with filters
	totalCount, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count workflows: %w", err)
	}
	result.TotalCount = totalCount

	// Fetch records with pagination
	var workflows []WorkflowMeta
	err = db.NewSelect().
		Model(&workflows).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			// Filter out hidden workflows by default
			q = q.Where("hidden = ? OR hidden IS NULL", false)
			if query.Kind != "" {
				q = q.Where("kind = ?", query.Kind)
			}
			if query.Search != "" {
				searchPattern := "%" + query.Search + "%"
				q = q.Where("(name LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
			}
			if len(query.Tags) > 0 {
				q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
					for _, tag := range query.Tags {
						if IsSQLite() {
							sq = sq.WhereOr("EXISTS (SELECT 1 FROM json_each(tags) WHERE value = ?)", tag)
						} else {
							sq = sq.WhereOr("tags @> ?", fmt.Sprintf(`["%s"]`, tag))
						}
					}
					return sq
				})
			}
			return q
		}).
		Order("name ASC").
		Offset(result.Offset).
		Limit(result.Limit).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch workflows: %w", err)
	}

	result.Data = workflows
	return result, nil
}

// GetWorkflowFromDB returns a single workflow metadata by name
func GetWorkflowFromDB(ctx context.Context, name string) (*WorkflowMeta, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Try cache first
	if cache := GetCache(); cache != nil {
		if meta, found := cache.GetWorkflowMeta(name); found {
			return meta, nil
		}
	}

	// Cache miss - query database
	var meta WorkflowMeta
	err := db.NewSelect().Model(&meta).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if cache := GetCache(); cache != nil {
		cache.SetWorkflowMeta(name, &meta)
	}

	return &meta, nil
}

// GetAllTags returns all unique tags from workflows
func GetAllTags(ctx context.Context) ([]string, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var workflows []WorkflowMeta
	if err := db.NewSelect().Model(&workflows).Column("tags").Scan(ctx); err != nil {
		return nil, err
	}

	// Collect unique tags
	tagMap := make(map[string]bool)
	for _, w := range workflows {
		for _, tag := range w.Tags {
			tagMap[strings.TrimSpace(tag)] = true
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// GetWorkflowCount returns the total number of indexed workflows
func GetWorkflowCount(ctx context.Context) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("database not connected")
	}

	return db.NewSelect().Model((*WorkflowMeta)(nil)).Count(ctx)
}
