package updater

import (
	"context"
	"strings"
	"time"
)

// MockSource is a mock implementation of Source for testing
type MockSource struct {
	Releases     []*Release
	DetectError  error
	UpdateError  error
	DetectCalled int
	UpdateCalled int
	LastUpdateTo *Release
}

// NewMockSource creates a mock source with predefined releases
func NewMockSource() *MockSource {
	return &MockSource{
		Releases: []*Release{},
	}
}

// AddRelease adds a release to the mock
func (m *MockSource) AddRelease(rel *Release) {
	m.Releases = append(m.Releases, rel)
}

// SetLatest sets the latest release
func (m *MockSource) SetLatest(version, assetURL string, prerelease bool) {
	m.Releases = append(m.Releases, &Release{
		Version:     version,
		AssetURL:    assetURL,
		Prerelease:  prerelease,
		PublishedAt: time.Now(),
	})
}

// SetError sets the error to return from DetectLatest/DetectVersion
func (m *MockSource) SetError(err error) {
	m.DetectError = err
}

// SetUpdateError sets the error to return from UpdateTo
func (m *MockSource) SetUpdateError(err error) {
	m.UpdateError = err
}

func (m *MockSource) DetectLatest(ctx context.Context, owner, repo string) (*Release, error) {
	m.DetectCalled++
	if m.DetectError != nil {
		return nil, m.DetectError
	}

	// Find the latest non-prerelease
	var latest *Release
	for _, rel := range m.Releases {
		if rel.Prerelease {
			continue
		}
		if latest == nil {
			latest = rel
			continue
		}
		isNewer, err := IsNewerVersion(rel.Version, latest.Version)
		if err == nil && isNewer {
			latest = rel
		}
	}
	return latest, nil
}

func (m *MockSource) DetectVersion(ctx context.Context, owner, repo, version string) (*Release, error) {
	m.DetectCalled++
	if m.DetectError != nil {
		return nil, m.DetectError
	}

	// Find matching version
	version = strings.TrimPrefix(version, "v")
	for _, rel := range m.Releases {
		relVersion := strings.TrimPrefix(rel.Version, "v")
		if relVersion == version {
			return rel, nil
		}
	}
	return nil, nil
}

func (m *MockSource) UpdateTo(ctx context.Context, owner, repo string, release *Release) error {
	m.UpdateCalled++
	m.LastUpdateTo = release
	if m.UpdateError != nil {
		return m.UpdateError
	}
	return nil
}
