//go:build windows

package fileio

import (
	"os"
	"syscall"
	"unsafe"
)

// mmapFile memory-maps a file for reading on Windows systems.
func mmapFile(f *os.File, size int64) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}

	// Create a file mapping object
	low := uint32(size)
	high := uint32(size >> 32)
	h, err := syscall.CreateFileMapping(syscall.Handle(f.Fd()), nil, syscall.PAGE_READONLY, high, low, nil)
	if err != nil {
		return nil, err
	}

	// Map the file into memory
	ptr, err := syscall.MapViewOfFile(h, syscall.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		syscall.CloseHandle(h)
		return nil, err
	}

	// Note: We don't close h here because we need to keep the mapping alive
	// The handle will be closed when the process exits or when we unmap

	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size), nil
}

// munmapFile unmaps a memory-mapped file on Windows systems.
func munmapFile(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return syscall.UnmapViewOfFile(uintptr(unsafe.Pointer(&data[0])))
}
