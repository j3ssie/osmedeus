package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ErrOrgNotFound is returned when an org reference resolves to nothing.
var ErrOrgNotFound = errors.New("org not found")

// ErrDefaultOrgImmutable is returned when a caller tries to delete or rename the
// built-in default org.
var ErrDefaultOrgImmutable = errors.New("the default org cannot be deleted or renamed")

// NormalizeOrgUUID coerces an empty org UUID to the default org.
//
// Call this at every write boundary *before* any dedup lookup. Bun persists the Go
// zero value rather than falling back to the column DEFAULT, so an unset OrgUUID
// would be stored as an empty string while a subsequent lookup filtering on the
// default UUID would miss it - letting duplicates through and hiding the row from
// every org query.
func NormalizeOrgUUID(orgUUID string) string {
	if strings.TrimSpace(orgUUID) == "" {
		return DefaultOrgUUID
	}
	return orgUUID
}

// CreateOrg inserts a new org. If explicitUUID is empty a v4 UUID is generated.
func CreateOrg(ctx context.Context, name, description, explicitUUID string, tags []string) (*Org, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("org name cannot be empty")
	}

	orgUUID := strings.TrimSpace(explicitUUID)
	if orgUUID == "" {
		orgUUID = uuid.New().String()
	}

	now := time.Now()
	org := &Org{
		UUID:        orgUUID,
		Name:        name,
		Description: description,
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if _, err := db.NewInsert().Model(org).Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to create org %q: %w", name, err)
	}
	return org, nil
}

// GetOrgByUUID looks up an org by its UUID.
func GetOrgByUUID(ctx context.Context, orgUUID string) (*Org, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	org := new(Org)
	err := db.NewSelect().Model(org).Where("uuid = ?", orgUUID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %s", ErrOrgNotFound, orgUUID)
	}
	if err != nil {
		return nil, err
	}
	return org, nil
}

// GetOrgByName looks up an org by its unique name (case-insensitive).
func GetOrgByName(ctx context.Context, name string) (*Org, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	org := new(Org)
	err := db.NewSelect().Model(org).
		Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %s", ErrOrgNotFound, name)
	}
	if err != nil {
		return nil, err
	}
	return org, nil
}

// ResolveOrgRef resolves a user-supplied reference that may be either an org name
// or an org UUID. A well-formed UUID is tried as a UUID first and anything else as
// a name first, but both lookups are attempted either way so a name that happens to
// look like a UUID still resolves.
func ResolveOrgRef(ctx context.Context, ref string) (*Org, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("%w: empty reference", ErrOrgNotFound)
	}

	first, second := GetOrgByName, GetOrgByUUID
	if _, err := uuid.Parse(ref); err == nil {
		first, second = GetOrgByUUID, GetOrgByName
	}

	org, err := first(ctx, ref)
	if err == nil || !errors.Is(err, ErrOrgNotFound) {
		return org, err
	}
	return second(ctx, ref)
}

// ResolveOrgUUID resolves a name-or-UUID reference to a UUID. An empty reference
// resolves to an empty string, which read paths treat as "no org filter".
func ResolveOrgUUID(ctx context.Context, ref string) (string, error) {
	if strings.TrimSpace(ref) == "" {
		return "", nil
	}
	org, err := ResolveOrgRef(ctx, ref)
	if err != nil {
		return "", err
	}
	return org.UUID, nil
}

// ListOrgs returns every org ordered by name, default org first.
func ListOrgs(ctx context.Context) ([]Org, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var orgs []Org
	err := db.NewSelect().Model(&orgs).
		OrderExpr("CASE WHEN uuid = ? THEN 0 ELSE 1 END", DefaultOrgUUID).
		Order("name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

// UpdateOrg updates an org's name, description and tags. Empty name or description
// leaves the existing value untouched; pass tags as non-nil to replace them.
func UpdateOrg(ctx context.Context, orgUUID, name, description string, tags []string) (*Org, error) {
	org, err := GetOrgByUUID(ctx, orgUUID)
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name != "" && !strings.EqualFold(name, org.Name) {
		if org.IsDefault() {
			return nil, ErrDefaultOrgImmutable
		}
		org.Name = name
	}
	if description != "" {
		org.Description = description
	}
	if tags != nil {
		org.Tags = tags
	}
	org.UpdatedAt = time.Now()

	_, err = db.NewUpdate().Model(org).
		Column("name", "description", "tags", "updated_at").
		WherePK().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update org: %w", err)
	}
	return org, nil
}

// DeleteOrg removes an org.
//
// With purge=false the org's data is reassigned to the default org — nothing is
// lost, the grouping is simply dropped. With purge=true every workspace, asset,
// vulnerability and run belonging to the org is deleted along with it.
//
// The default org can never be deleted.
func DeleteOrg(ctx context.Context, orgUUID string, purge bool) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}
	if orgUUID == DefaultOrgUUID {
		return ErrDefaultOrgImmutable
	}
	if _, err := GetOrgByUUID(ctx, orgUUID); err != nil {
		return err
	}

	return Transaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		for _, table := range orgScopedTables {
			var err error
			if purge {
				_, err = tx.NewDelete().Table(table).Where("org_uuid = ?", orgUUID).Exec(ctx)
			} else {
				_, err = tx.NewUpdate().Table(table).
					Set("org_uuid = ?", DefaultOrgUUID).
					Where("org_uuid = ?", orgUUID).Exec(ctx)
			}
			if err != nil {
				return fmt.Errorf("failed to clear org data from %s: %w", table, err)
			}
		}

		if _, err := tx.NewDelete().Model((*Org)(nil)).Where("uuid = ?", orgUUID).Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete org: %w", err)
		}
		invalidateWorkspaceOrgCache()
		return nil
	})
}

// AssignWorkspacesToOrg moves whole workspaces into an org, cascading the stamp to
// every asset, vulnerability and run belonging to those workspaces. This is how
// pre-existing data gets grouped without re-scanning.
//
// Returns the number of rows updated per table, keyed by table name.
func AssignWorkspacesToOrg(ctx context.Context, orgUUID string, workspaces []string) (map[string]int64, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	if len(workspaces) == 0 {
		return nil, fmt.Errorf("no workspaces given")
	}
	if _, err := GetOrgByUUID(ctx, orgUUID); err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(orgScopedTables))
	err := Transaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		for _, table := range orgScopedTables {
			// The workspaces table keys on `name`; every other org-scoped table
			// carries a denormalized `workspace` column.
			column := "workspace"
			if table == "workspaces" {
				column = "name"
			}

			res, err := tx.NewUpdate().
				Table(table).
				Set("org_uuid = ?", orgUUID).
				Where(column+" IN (?)", bun.In(workspaces)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to assign org on %s: %w", table, err)
			}
			if n, err := res.RowsAffected(); err == nil {
				counts[table] = n
			}
		}
		invalidateWorkspaceOrgCache()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return counts, nil
}

// GetOrgStats returns aggregate counts for an org.
func GetOrgStats(ctx context.Context, orgUUID string) (*OrgStats, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	org, err := GetOrgByUUID(ctx, orgUUID)
	if err != nil {
		return nil, err
	}

	stats := &OrgStats{OrgUUID: org.UUID, OrgName: org.Name}

	if stats.TotalAssets, err = db.NewSelect().Model((*Asset)(nil)).
		Where("org_uuid = ?", orgUUID).Count(ctx); err != nil {
		return nil, err
	}
	if stats.TotalVulns, err = db.NewSelect().Model((*Vulnerability)(nil)).
		Where("org_uuid = ?", orgUUID).Count(ctx); err != nil {
		return nil, err
	}
	if stats.TotalRuns, err = db.NewSelect().Model((*Run)(nil)).
		Where("org_uuid = ?", orgUUID).Count(ctx); err != nil {
		return nil, err
	}

	if err := db.NewSelect().Model((*Workspace)(nil)).
		Column("name").
		Where("org_uuid = ?", orgUUID).
		Order("name ASC").
		Scan(ctx, &stats.Workspaces); err != nil {
		return nil, err
	}
	stats.TotalWorkspaces = len(stats.Workspaces)

	return stats, nil
}
