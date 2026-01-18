package updater

import (
	"context"
	"fmt"

	selfupdate "github.com/creativeprojects/go-selfupdate"
)

// GitHubSource implements Source using go-selfupdate's GitHub backend
type GitHubSource struct {
	updater         *selfupdate.Updater
	allowPrerelease bool
}

// NewGitHubSource creates a new GitHub release source
func NewGitHubSource(allowPrerelease bool) (*GitHubSource, error) {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub source: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:     source,
		Prerelease: allowPrerelease,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create updater: %w", err)
	}

	return &GitHubSource{
		updater:         updater,
		allowPrerelease: allowPrerelease,
	}, nil
}

func (g *GitHubSource) DetectLatest(ctx context.Context, owner, repo string) (*Release, error) {
	repository := selfupdate.NewRepositorySlug(owner, repo)
	rel, found, err := g.updater.DetectLatest(ctx, repository)
	if err != nil {
		return nil, fmt.Errorf("failed to detect latest release: %w", err)
	}
	if !found {
		return nil, nil
	}
	return convertSelfupdateRelease(rel), nil
}

func (g *GitHubSource) DetectVersion(ctx context.Context, owner, repo, version string) (*Release, error) {
	repository := selfupdate.NewRepositorySlug(owner, repo)
	rel, found, err := g.updater.DetectVersion(ctx, repository, version)
	if err != nil {
		return nil, fmt.Errorf("failed to detect version %s: %w", version, err)
	}
	if !found {
		return nil, nil
	}
	return convertSelfupdateRelease(rel), nil
}

func (g *GitHubSource) UpdateTo(ctx context.Context, owner, repo string, release *Release) error {
	// Get the executable path
	exe, err := GetExecutablePath()
	if err != nil {
		return err
	}

	// Detect the release again to get the full selfupdate.Release object
	repository := selfupdate.NewRepositorySlug(owner, repo)
	rel, found, err := g.updater.DetectVersion(ctx, repository, release.Version)
	if err != nil {
		return fmt.Errorf("failed to detect release for update: %w", err)
	}
	if !found {
		return fmt.Errorf("release %s not found", release.Version)
	}

	// Perform the update
	if err := g.updater.UpdateTo(ctx, rel, exe); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	return nil
}
