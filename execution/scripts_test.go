package execution

import (
	"testing"
)

func TestChunkFile(t *testing.T) {
	result := ChunkFileByPart("/tmp/oo/seqtest", 3)
	t.Log(result)
}

func TestSort(t *testing.T) {
	Sort("/tmp/sam")
}

func TestIsWildCard(t *testing.T) {
	IsWildCard("github.com")
	IsWildCard("tesla.com")
}
