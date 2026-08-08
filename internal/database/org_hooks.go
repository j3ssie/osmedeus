package database

import (
	"context"
	"sync"

	"github.com/uptrace/bun"
)

// The BeforeAppendModel hooks below are the last line of defence for org
// attribution. Bun persists the Go zero value rather than falling back to the
// column DEFAULT, so any write that forgets to set OrgUUID would store an empty
// string - a value that matches no org filter and hides the row from every query.
//
// The hooks fire on INSERT and UPDATE alike, because the import paths build a
// fresh struct from scan output and UPDATE it WherePK; such a struct has an empty
// OrgUUID meaning "not loaded", not "default org". Resolving from the row's
// workspace produces the right answer in both cases, since org membership is
// assigned per workspace and cascaded (see AssignWorkspacesToOrg).
//
// Writes that deliberately set org_uuid — `org assign`, `org delete` — use raw
// UPDATE statements without a model, so these hooks never fire on them.
//
// A non-empty OrgUUID is always left untouched.

var _ bun.BeforeAppendModelHook = (*Asset)(nil)
var _ bun.BeforeAppendModelHook = (*Vulnerability)(nil)
var _ bun.BeforeAppendModelHook = (*Run)(nil)
var _ bun.BeforeAppendModelHook = (*Workspace)(nil)

// orgForWrite returns the org a model should be persisted under: the one it
// already names, or the org of the workspace it belongs to.
//
// Passing an empty workspace yields the default org, which is what Workspace
// itself wants — it is the root of the mapping, with nothing to derive from.
func orgForWrite(ctx context.Context, query bun.Query, current, workspace string) string {
	if current != "" {
		return current
	}
	switch query.(type) {
	case *bun.InsertQuery, *bun.UpdateQuery:
		return GetOrgUUIDForWorkspace(ctx, workspace)
	}
	return current
}

func (a *Asset) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	a.OrgUUID = orgForWrite(ctx, query, a.OrgUUID, a.Workspace)
	return nil
}

func (v *Vulnerability) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	v.OrgUUID = orgForWrite(ctx, query, v.OrgUUID, v.Workspace)
	return nil
}

func (r *Run) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	r.OrgUUID = orgForWrite(ctx, query, r.OrgUUID, r.Workspace)
	return nil
}

func (w *Workspace) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	w.OrgUUID = orgForWrite(ctx, query, w.OrgUUID, "")
	return nil
}

// workspaceOrgCache memoizes workspace name -> org UUID so per-line importers do
// not issue a lookup per row. Invalidated wholesale whenever org membership
// changes, which is rare.
var (
	workspaceOrgCache   = make(map[string]string)
	workspaceOrgCacheMu sync.RWMutex
)

// invalidateWorkspaceOrgCache drops every memoized workspace -> org mapping.
func invalidateWorkspaceOrgCache() {
	workspaceOrgCacheMu.Lock()
	workspaceOrgCache = make(map[string]string)
	workspaceOrgCacheMu.Unlock()
}

// GetOrgUUIDForWorkspace returns the org a workspace belongs to, falling back to
// the default org when the workspace has no row yet (the first import of a scan
// can land before EnsureWorkspaceRuntime).
//
// Deriving the org from the workspace rather than threading it through every
// import function means org membership always follows the workspace: reassigning
// a workspace with `osmedeus org assign` retargets its future imports too.
func GetOrgUUIDForWorkspace(ctx context.Context, workspace string) string {
	if workspace == "" {
		return DefaultOrgUUID
	}

	workspaceOrgCacheMu.RLock()
	cached, ok := workspaceOrgCache[workspace]
	workspaceOrgCacheMu.RUnlock()
	if ok {
		return cached
	}

	if db == nil {
		return DefaultOrgUUID
	}

	var orgUUID string
	err := db.NewSelect().Model((*Workspace)(nil)).
		Column("org_uuid").
		Where("name = ?", workspace).
		Limit(1).
		Scan(ctx, &orgUUID)
	if err != nil || orgUUID == "" {
		// Do not cache a miss: the workspace row usually appears moments later.
		return DefaultOrgUUID
	}

	workspaceOrgCacheMu.Lock()
	workspaceOrgCache[workspace] = orgUUID
	workspaceOrgCacheMu.Unlock()
	return orgUUID
}
