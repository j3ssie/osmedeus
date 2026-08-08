package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// activeOrgFileName is the file under the base folder holding the org selected by
// `osmedeus org use`.
const activeOrgFileName = ".active-org"

var (
	resolvedOrgUUID string
	resolveOrgOnce  sync.Once
	resolveOrgErr   error
)

// activeOrgFilePath returns the path of the persisted active-org file.
func activeOrgFilePath() string {
	base := ""
	if cfg := config.Get(); cfg != nil {
		base = cfg.BaseFolder
	}
	if base == "" {
		base = baseFolder
	}
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, "osmedeus-base")
	}
	return filepath.Join(base, activeOrgFileName)
}

// readActiveOrg returns the persisted active org reference, or "" if none is set.
func readActiveOrg() string {
	path := activeOrgFilePath()
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// writeActiveOrg persists the active org reference.
func writeActiveOrg(orgUUID string) error {
	path := activeOrgFilePath()
	if path == "" {
		return fmt.Errorf("cannot determine base folder for active org file")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(orgUUID+"\n"), 0o644)
}

// clearActiveOrg removes the persisted active org, returning to the unfiltered
// "all orgs" view.
func clearActiveOrg() error {
	path := activeOrgFilePath()
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// orgRef returns the raw org reference (name or UUID) from the highest-priority
// source that has one, without touching the database. Resolution order:
//
//  1. --org flag
//  2. $OSMEDEUS_ORG_UUID
//  3. $OSMEDEUS_ORG
//  4. the active-org file written by `osmedeus org use`
//  5. "" — meaning no org was selected
func orgRef() string {
	if v := strings.TrimSpace(globalOrg); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("OSMEDEUS_ORG_UUID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("OSMEDEUS_ORG")); v != "" {
		return v
	}
	return readActiveOrg()
}

// resolveOrgUUID resolves the selected org to a UUID, once per process.
//
// The memo matters: a multi-target run resolves the org once per target, and all
// targets share one constant reference.
//
// An empty return means no org was selected. Read paths must treat that as "no
// filter" so a database with no orgs behaves exactly as it did before the org
// layer existed; write paths must run it through database.NormalizeOrgUUID so
// rows land in the default org rather than being stranded under an empty string.
//
// Requires a connected database; callers that may run without one should check
// database.GetDB() first.
func resolveOrgUUID(ctx context.Context) (string, error) {
	resolveOrgOnce.Do(func() {
		ref := orgRef()
		if ref == "" {
			return
		}
		orgUUID, err := database.ResolveOrgUUID(ctx, ref)
		if err != nil {
			resolveOrgErr = fmt.Errorf("failed to resolve --org %q: %w", ref, err)
			return
		}
		resolvedOrgUUID = orgUUID
	})
	return resolvedOrgUUID, resolveOrgErr
}

// resetOrgResolution clears the cached resolution. Tests only.
func resetOrgResolution() {
	resolveOrgOnce = sync.Once{}
	resolvedOrgUUID = ""
	resolveOrgErr = nil
}
