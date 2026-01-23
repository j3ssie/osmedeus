//go:build unix

package fileio

import (
	"os"
	"syscall"
)

// mmapFile memory-maps a file for reading on Unix systems.
func mmapFile(f *os.File, size int64) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}

	data, err := syscall.Mmap(
		int(f.Fd()),
		0,
		int(size),
		syscall.PROT_READ,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// munmapFile unmaps a memory-mapped file on Unix systems.
func munmapFile(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return syscall.Munmap(data)
}
