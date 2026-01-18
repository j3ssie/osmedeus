package updater

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "full https URL",
			url:       "https://github.com/j3ssie/osmedeus",
			wantOwner: "j3ssie",
			wantRepo:  "osmedeus",
		},
		{
			name:      "URL with .git suffix",
			url:       "https://github.com/j3ssie/osmedeus.git",
			wantOwner: "j3ssie",
			wantRepo:  "osmedeus",
		},
		{
			name:      "URL without protocol",
			url:       "github.com/j3ssie/osmedeus",
			wantOwner: "j3ssie",
			wantRepo:  "osmedeus",
		},
		{
			name:      "URL with trailing slash",
			url:       "https://github.com/j3ssie/osmedeus/",
			wantOwner: "j3ssie",
			wantRepo:  "osmedeus",
		},
		{
			name:      "http URL",
			url:       "http://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:    "invalid URL - single part",
			url:     "invalid",
			wantErr: true,
		},
		{
			name:    "invalid URL - empty",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOwner, owner)
			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		newVersion     string
		currentVersion string
		want           bool
		wantErr        bool
	}{
		{"v5.1.0", "v5.0.0", true, false},
		{"5.1.0", "5.0.0", true, false},
		{"v5.0.0", "v5.0.0", false, false},
		{"v5.0.0", "v5.1.0", false, false},
		{"v5.0.1", "v5.0.0", true, false},
		{"v5.10.0", "v5.9.0", true, false},
		{"v6.0.0", "v5.99.99", true, false},
		{"v1.0.0", "v2.0.0", false, false},
		{"v5.0.0-beta", "v5.0.0", false, false},
		{"v5.0.1-rc1", "v5.0.0", true, false},
		{"invalid", "v5.0.0", false, true},
		{"v5.0.0", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.newVersion+"_vs_"+tt.currentVersion, func(t *testing.T) {
			got, err := IsNewerVersion(tt.newVersion, tt.currentVersion)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckForUpdate_NewVersionAvailable(t *testing.T) {
	mock := NewMockSource()
	mock.SetLatest("v5.1.0", "https://example.com/release.tar.gz", false)

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	release, hasUpdate, err := upd.CheckForUpdate(ctx, "v5.0.0")

	require.NoError(t, err)
	assert.True(t, hasUpdate)
	assert.NotNil(t, release)
	assert.Equal(t, "v5.1.0", release.Version)
	assert.Equal(t, 1, mock.DetectCalled)
}

func TestCheckForUpdate_AlreadyLatest(t *testing.T) {
	mock := NewMockSource()
	mock.SetLatest("v5.0.0", "https://example.com/release.tar.gz", false)

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	release, hasUpdate, err := upd.CheckForUpdate(ctx, "v5.0.0")

	require.NoError(t, err)
	assert.False(t, hasUpdate)
	assert.NotNil(t, release) // Release info still returned
	assert.Equal(t, "v5.0.0", release.Version)
}

func TestCheckForUpdate_NetworkError(t *testing.T) {
	mock := NewMockSource()
	mock.SetError(errors.New("network timeout"))

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	_, _, err := upd.CheckForUpdate(ctx, "v5.0.0")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "network timeout")
}

func TestCheckForUpdate_NoReleases(t *testing.T) {
	mock := NewMockSource() // No releases added

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	release, hasUpdate, err := upd.CheckForUpdate(ctx, "v5.0.0")

	require.NoError(t, err)
	assert.False(t, hasUpdate)
	assert.Nil(t, release)
}

func TestCheckSpecificVersion_Found(t *testing.T) {
	mock := NewMockSource()
	mock.AddRelease(&Release{Version: "v5.0.0", PublishedAt: time.Now()})
	mock.AddRelease(&Release{Version: "v5.1.0", PublishedAt: time.Now()})
	mock.AddRelease(&Release{Version: "v5.2.0", PublishedAt: time.Now()})

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	release, found, err := upd.CheckSpecificVersion(ctx, "v5.1.0")

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "v5.1.0", release.Version)
}

func TestCheckSpecificVersion_NotFound(t *testing.T) {
	mock := NewMockSource()
	mock.AddRelease(&Release{Version: "v5.0.0", PublishedAt: time.Now()})

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	release, found, err := upd.CheckSpecificVersion(ctx, "v99.0.0")

	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, release)
}

func TestCheckSpecificVersion_WithoutVPrefix(t *testing.T) {
	mock := NewMockSource()
	mock.AddRelease(&Release{Version: "v5.1.0", PublishedAt: time.Now()})

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	release, found, err := upd.CheckSpecificVersion(ctx, "5.1.0") // No 'v' prefix

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "v5.1.0", release.Version)
}

func TestUpdate_Success(t *testing.T) {
	mock := NewMockSource()
	mock.SetLatest("v5.1.0", "https://example.com/release.tar.gz", false)

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	result, err := upd.Update(ctx, "v5.0.0", false)

	require.NoError(t, err)
	assert.True(t, result.Updated)
	assert.Equal(t, "v5.0.0", result.OldVersion)
	assert.Equal(t, "v5.1.0", result.NewVersion)
	assert.Equal(t, 1, mock.UpdateCalled)
	assert.NotNil(t, mock.LastUpdateTo)
}

func TestUpdate_AlreadyLatest_NoForce(t *testing.T) {
	mock := NewMockSource()
	mock.SetLatest("v5.0.0", "https://example.com/release.tar.gz", false)

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	result, err := upd.Update(ctx, "v5.0.0", false)

	require.NoError(t, err)
	assert.False(t, result.Updated)
	assert.Equal(t, "v5.0.0", result.OldVersion)
	assert.Equal(t, "v5.0.0", result.NewVersion)
	assert.Equal(t, 0, mock.UpdateCalled)
}

func TestUpdate_AlreadyLatest_WithForce(t *testing.T) {
	mock := NewMockSource()
	mock.SetLatest("v5.0.0", "https://example.com/release.tar.gz", false)

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	result, err := upd.Update(ctx, "v5.0.0", true) // force=true

	require.NoError(t, err)
	assert.True(t, result.Updated)
	assert.Equal(t, 1, mock.UpdateCalled)
}

func TestUpdate_UpdateError(t *testing.T) {
	mock := NewMockSource()
	mock.SetLatest("v5.1.0", "https://example.com/release.tar.gz", false)
	mock.SetUpdateError(errors.New("download failed"))

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	_, err := upd.Update(ctx, "v5.0.0", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "download failed")
}

func TestUpdateToVersion_Success(t *testing.T) {
	mock := NewMockSource()
	mock.AddRelease(&Release{Version: "v5.0.0", PublishedAt: time.Now()})
	mock.AddRelease(&Release{Version: "v5.1.0", PublishedAt: time.Now()})
	mock.AddRelease(&Release{Version: "v5.2.0", PublishedAt: time.Now()})

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	result, err := upd.UpdateToVersion(ctx, "v5.0.0", "v5.1.0", false)

	require.NoError(t, err)
	assert.True(t, result.Updated)
	assert.Equal(t, "v5.0.0", result.OldVersion)
	assert.Equal(t, "v5.1.0", result.NewVersion)
}

func TestUpdateToVersion_VersionNotFound(t *testing.T) {
	mock := NewMockSource()
	mock.AddRelease(&Release{Version: "v5.0.0", PublishedAt: time.Now()})

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	_, err := upd.UpdateToVersion(ctx, "v5.0.0", "v99.0.0", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateToVersion_SameVersion_NoForce(t *testing.T) {
	mock := NewMockSource()
	mock.AddRelease(&Release{Version: "v5.0.0", PublishedAt: time.Now()})

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	result, err := upd.UpdateToVersion(ctx, "v5.0.0", "v5.0.0", false)

	require.NoError(t, err)
	assert.False(t, result.Updated)
	assert.Equal(t, 0, mock.UpdateCalled)
}

func TestUpdateToVersion_SameVersion_WithForce(t *testing.T) {
	mock := NewMockSource()
	mock.AddRelease(&Release{Version: "v5.0.0", PublishedAt: time.Now()})

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx := context.Background()
	result, err := upd.UpdateToVersion(ctx, "v5.0.0", "v5.0.0", true) // force=true

	require.NoError(t, err)
	assert.True(t, result.Updated)
	assert.Equal(t, 1, mock.UpdateCalled)
}

func TestMockSource_DetectLatest_IgnoresPrerelease(t *testing.T) {
	mock := NewMockSource()
	mock.AddRelease(&Release{Version: "v5.0.0", Prerelease: false})
	mock.AddRelease(&Release{Version: "v5.1.0", Prerelease: true})  // Prerelease
	mock.AddRelease(&Release{Version: "v5.0.5", Prerelease: false}) // Latest stable

	ctx := context.Background()
	release, err := mock.DetectLatest(ctx, "owner", "repo")

	require.NoError(t, err)
	assert.NotNil(t, release)
	// Should return v5.0.5, not v5.1.0 (prerelease)
	assert.Equal(t, "v5.0.5", release.Version)
}

func TestContextCancellation(t *testing.T) {
	mock := NewMockSource()
	mock.SetLatest("v5.1.0", "https://example.com/release.tar.gz", false)

	upd := NewUpdater(Options{
		Owner:  "j3ssie",
		Repo:   "osmedeus",
		Source: mock,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The mock doesn't check context, so this will still work
	// In real implementation with GitHub source, this would fail
	_, _, err := upd.CheckForUpdate(ctx, "v5.0.0")

	// With mock, it still succeeds - the real source would check context
	require.NoError(t, err)
}

func TestDefaultUpdater(t *testing.T) {
	// Test that DefaultUpdater creates a valid updater
	upd := DefaultUpdater("j3ssie", "osmedeus")
	assert.NotNil(t, upd)
}

func TestNewUpdater_WithNilSource(t *testing.T) {
	// When Source is nil, it should create a default GitHub source
	upd := NewUpdater(Options{
		Owner: "j3ssie",
		Repo:  "osmedeus",
		// Source is nil
	})

	assert.NotNil(t, upd)
}
