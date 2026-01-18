package updater

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	selfupdate "github.com/creativeprojects/go-selfupdate"
)

// Release represents a software release
type Release struct {
	Version      string
	AssetURL     string
	AssetName    string
	AssetSize    int
	ReleaseNotes string
	PublishedAt  time.Time
	Prerelease   bool
}

// UpdateResult contains the result of an update operation
type UpdateResult struct {
	Updated      bool
	OldVersion   string
	NewVersion   string
	ReleaseNotes string
}

// Source is the interface for fetching releases (abstracts go-selfupdate Source)
type Source interface {
	// DetectLatest returns the latest release, or nil if none found
	DetectLatest(ctx context.Context, owner, repo string) (*Release, error)

	// DetectVersion returns a specific version release
	DetectVersion(ctx context.Context, owner, repo, version string) (*Release, error)

	// UpdateTo updates to the specified release
	UpdateTo(ctx context.Context, owner, repo string, release *Release) error
}

// Updater is the interface for self-update operations
type Updater interface {
	// CheckForUpdate checks if a newer version is available
	CheckForUpdate(ctx context.Context, currentVersion string) (*Release, bool, error)

	// CheckSpecificVersion checks if a specific version is available
	CheckSpecificVersion(ctx context.Context, version string) (*Release, bool, error)

	// Update performs the self-update to the latest version
	Update(ctx context.Context, currentVersion string, force bool) (*UpdateResult, error)

	// UpdateToVersion updates to a specific version
	UpdateToVersion(ctx context.Context, currentVersion, targetVersion string, force bool) (*UpdateResult, error)
}

// Options for configuring the updater
type Options struct {
	// Owner is the GitHub repository owner (e.g., "j3ssie")
	Owner string

	// Repo is the GitHub repository name (e.g., "osmedeus")
	Repo string

	// AllowPrerelease allows updating to prerelease versions
	AllowPrerelease bool

	// Source is the release source (defaults to GitHub)
	Source Source
}

// osmUpdater implements the Updater interface
type osmUpdater struct {
	owner           string
	repo            string
	source          Source
	allowPrerelease bool
}

// NewUpdater creates an updater with custom options
func NewUpdater(opts Options) Updater {
	source := opts.Source
	if source == nil {
		// Default to GitHub source
		githubSource, err := NewGitHubSource(opts.AllowPrerelease)
		if err != nil {
			// If we can't create GitHub source, use a nil source
			// This will cause errors when trying to update, but allows construction
			source = nil
		} else {
			source = githubSource
		}
	}

	return &osmUpdater{
		owner:           opts.Owner,
		repo:            opts.Repo,
		source:          source,
		allowPrerelease: opts.AllowPrerelease,
	}
}

// DefaultUpdater creates an updater with default options
func DefaultUpdater(owner, repo string) Updater {
	return NewUpdater(Options{
		Owner: owner,
		Repo:  repo,
	})
}

func (u *osmUpdater) CheckForUpdate(ctx context.Context, currentVersion string) (*Release, bool, error) {
	if u.source == nil {
		return nil, false, fmt.Errorf("no release source configured")
	}

	release, err := u.source.DetectLatest(ctx, u.owner, u.repo)
	if err != nil {
		return nil, false, fmt.Errorf("failed to detect latest release: %w", err)
	}
	if release == nil {
		return nil, false, nil
	}

	// Compare versions using semver
	hasUpdate, err := IsNewerVersion(release.Version, currentVersion)
	if err != nil {
		// If semver comparison fails, fall back to string comparison
		hasUpdate = release.Version != currentVersion
	}

	return release, hasUpdate, nil
}

func (u *osmUpdater) CheckSpecificVersion(ctx context.Context, version string) (*Release, bool, error) {
	if u.source == nil {
		return nil, false, fmt.Errorf("no release source configured")
	}

	release, err := u.source.DetectVersion(ctx, u.owner, u.repo, version)
	if err != nil {
		return nil, false, fmt.Errorf("failed to detect version %s: %w", version, err)
	}
	if release == nil {
		return nil, false, nil
	}

	return release, true, nil
}

func (u *osmUpdater) Update(ctx context.Context, currentVersion string, force bool) (*UpdateResult, error) {
	release, hasUpdate, err := u.CheckForUpdate(ctx, currentVersion)
	if err != nil {
		return nil, err
	}

	if !hasUpdate && !force {
		return &UpdateResult{
			Updated:    false,
			OldVersion: currentVersion,
			NewVersion: currentVersion,
		}, nil
	}

	if release == nil {
		return nil, fmt.Errorf("no release available")
	}

	// Perform the actual update
	if err := u.source.UpdateTo(ctx, u.owner, u.repo, release); err != nil {
		return nil, fmt.Errorf("failed to apply update: %w", err)
	}

	return &UpdateResult{
		Updated:      true,
		OldVersion:   currentVersion,
		NewVersion:   release.Version,
		ReleaseNotes: release.ReleaseNotes,
	}, nil
}

func (u *osmUpdater) UpdateToVersion(ctx context.Context, currentVersion, targetVersion string, force bool) (*UpdateResult, error) {
	release, found, err := u.CheckSpecificVersion(ctx, targetVersion)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("version %s not found", targetVersion)
	}

	// Check if we're already on this version
	isSameVersion := strings.TrimPrefix(currentVersion, "v") == strings.TrimPrefix(release.Version, "v")
	if isSameVersion && !force {
		return &UpdateResult{
			Updated:    false,
			OldVersion: currentVersion,
			NewVersion: currentVersion,
		}, nil
	}

	// Perform the actual update
	if err := u.source.UpdateTo(ctx, u.owner, u.repo, release); err != nil {
		return nil, fmt.Errorf("failed to apply update: %w", err)
	}

	return &UpdateResult{
		Updated:      true,
		OldVersion:   currentVersion,
		NewVersion:   release.Version,
		ReleaseNotes: release.ReleaseNotes,
	}, nil
}

// ParseRepoURL extracts owner and repo from a GitHub URL
// Supports: https://github.com/owner/repo, github.com/owner/repo
func ParseRepoURL(repoURL string) (owner, repo string, err error) {
	// Remove protocol prefix
	url := strings.TrimPrefix(repoURL, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimSuffix(url, "/")

	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository URL: %s", repoURL)
	}
	return parts[0], parts[1], nil
}

// IsNewerVersion compares two semantic versions
// Returns true if new is greater than current
func IsNewerVersion(newVersion, currentVersion string) (bool, error) {
	// Strip 'v' prefix if present
	newVersion = strings.TrimPrefix(newVersion, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	newVer, err := semver.NewVersion(newVersion)
	if err != nil {
		return false, fmt.Errorf("invalid new version %s: %w", newVersion, err)
	}

	currentVer, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("invalid current version %s: %w", currentVersion, err)
	}

	return newVer.GreaterThan(currentVer), nil
}

// GetExecutablePath returns the path to the current executable
func GetExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}
	return exe, nil
}

// convertSelfupdateRelease converts a go-selfupdate Release to our Release type
func convertSelfupdateRelease(rel *selfupdate.Release) *Release {
	if rel == nil {
		return nil
	}
	return &Release{
		Version:      rel.Version(),
		AssetURL:     rel.AssetURL,
		AssetName:    rel.AssetName,
		AssetSize:    rel.AssetByteSize,
		ReleaseNotes: rel.ReleaseNotes,
		PublishedAt:  rel.PublishedAt,
		Prerelease:   rel.Prerelease,
	}
}
