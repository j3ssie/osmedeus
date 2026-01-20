package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkTargets(t *testing.T) {
	tests := []struct {
		name        string
		targets     []string
		size        int
		part        int
		wantTargets []string
		wantIndex   int
		wantTotal   int
		wantErr     string
	}{
		{
			name:        "no chunking when size=0",
			targets:     []string{"a", "b", "c"},
			size:        0,
			part:        0,
			wantTargets: []string{"a", "b", "c"},
		},
		{
			name:        "first chunk of 3",
			targets:     []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
			size:        3,
			part:        0,
			wantTargets: []string{"a", "b", "c"},
			wantIndex:   0,
			wantTotal:   3,
		},
		{
			name:        "middle chunk",
			targets:     []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
			size:        3,
			part:        1,
			wantTargets: []string{"d", "e", "f"},
			wantIndex:   1,
			wantTotal:   3,
		},
		{
			name:        "last chunk",
			targets:     []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
			size:        3,
			part:        2,
			wantTargets: []string{"g", "h", "i"},
			wantIndex:   2,
			wantTotal:   3,
		},
		{
			name:        "partial last chunk",
			targets:     []string{"a", "b", "c", "d", "e", "f", "g"},
			size:        3,
			part:        2,
			wantTargets: []string{"g"},
			wantIndex:   2,
			wantTotal:   3,
		},
		{
			name:    "chunk-part out of range",
			targets: []string{"a", "b", "c"},
			size:    3,
			part:    5,
			wantErr: "exceeds total chunks",
		},
		{
			name:    "info mode when part=-1",
			targets: []string{"a", "b", "c", "d", "e", "f"},
			size:    2,
			part:    -1,
			wantErr: "chunk-info",
		},
		{
			name:        "single target with chunking",
			targets:     []string{"only"},
			size:        10,
			part:        0,
			wantTargets: []string{"only"},
			wantIndex:   0,
			wantTotal:   1,
		},
		{
			name:        "empty targets",
			targets:     []string{},
			size:        10,
			part:        0,
			wantTargets: []string{},
		},
		{
			name:        "chunk larger than target count",
			targets:     []string{"a", "b", "c"},
			size:        10,
			part:        0,
			wantTargets: []string{"a", "b", "c"},
			wantIndex:   0,
			wantTotal:   1,
		},
		{
			name:        "exact division",
			targets:     []string{"a", "b", "c", "d", "e", "f"},
			size:        2,
			part:        2,
			wantTargets: []string{"e", "f"},
			wantIndex:   2,
			wantTotal:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, info, err := chunkTargets(tt.targets, tt.size, tt.part)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTargets, got)

			if info != nil {
				assert.Equal(t, tt.wantIndex, info.Index)
				assert.Equal(t, tt.wantTotal, info.Total)
			}
		})
	}
}

func TestChunkTargets_StartEnd(t *testing.T) {
	targets := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	got, info, err := chunkTargets(targets, 3, 1)
	require.NoError(t, err)

	assert.Equal(t, []string{"3", "4", "5"}, got)
	assert.Equal(t, 3, info.Start)
	assert.Equal(t, 6, info.End)
}

func TestChunkTargets_LastPartialChunk(t *testing.T) {
	// 10 targets with chunk size 3 = 4 chunks (3, 3, 3, 1)
	targets := make([]string, 10)
	for i := range targets {
		targets[i] = string(rune('a' + i))
	}

	// Last chunk should have only 1 target
	got, info, err := chunkTargets(targets, 3, 3)
	require.NoError(t, err)

	assert.Equal(t, 1, len(got))
	assert.Equal(t, "j", got[0])
	assert.Equal(t, 9, info.Start)
	assert.Equal(t, 10, info.End)
	assert.Equal(t, 4, info.Total)
}

func TestChunkTargets_InfoMode(t *testing.T) {
	targets := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	_, info, err := chunkTargets(targets, 4, -1)

	require.Error(t, err)
	assert.Equal(t, "chunk-info", err.Error())
	assert.NotNil(t, info)
	assert.Equal(t, 3, info.Total) // 10 targets / 4 = 3 chunks (4, 4, 2)
	assert.Equal(t, 4, info.Size)
}

func TestChunkInfo_Fields(t *testing.T) {
	targets := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	_, info, err := chunkTargets(targets, 4, 1)
	require.NoError(t, err)

	assert.Equal(t, 1, info.Index)
	assert.Equal(t, 4, info.Size)
	assert.Equal(t, 3, info.Total)
	assert.Equal(t, 4, info.Start)
	assert.Equal(t, 8, info.End)
}
