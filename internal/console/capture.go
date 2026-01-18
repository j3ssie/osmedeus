package console

import (
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Capture manages console output capture to file while maintaining terminal display
type Capture struct {
	mu          sync.Mutex
	file        *os.File
	originalOut *os.File
	originalErr *os.File
	outWriter   *os.File
	errWriter   *os.File
	outReader   *os.File
	errReader   *os.File
	done        chan struct{}
	wg          sync.WaitGroup
}

// StartCapture begins capturing stdout/stderr to the specified file
func StartCapture(filePath string) (*Capture, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, err
	}

	// Open file for writing
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	c := &Capture{
		file:        file,
		originalOut: os.Stdout,
		originalErr: os.Stderr,
		done:        make(chan struct{}),
	}

	// Create pipes for stdout and stderr
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		file.Close()
		return nil, err
	}

	errReader, errWriter, err := os.Pipe()
	if err != nil {
		file.Close()
		outReader.Close()
		outWriter.Close()
		return nil, err
	}

	c.outWriter = outWriter
	c.errWriter = errWriter
	c.outReader = outReader
	c.errReader = errReader

	// Replace stdout/stderr
	os.Stdout = outWriter
	os.Stderr = errWriter

	// Tee goroutines - write to both terminal and file
	c.wg.Add(2)
	go c.tee(outReader, c.originalOut)
	go c.tee(errReader, c.originalErr)

	return c, nil
}

func (c *Capture) tee(reader *os.File, terminal *os.File) {
	defer c.wg.Done()
	buf := make([]byte, 4096)
	for {
		select {
		case <-c.done:
			// Drain any remaining data
			c.drainReader(reader, terminal, buf)
			return
		default:
			n, err := reader.Read(buf)
			if n > 0 {
				data := buf[:n]
				_, _ = terminal.Write(data)
				c.mu.Lock()
				if c.file != nil {
					_, _ = c.file.Write(data)
				}
				c.mu.Unlock()
			}
			if err != nil {
				if err != io.EOF {
					return
				}
				return
			}
		}
	}
}

func (c *Capture) drainReader(reader *os.File, terminal *os.File, buf []byte) {
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			data := buf[:n]
			_, _ = terminal.Write(data)
			c.mu.Lock()
			if c.file != nil {
				_, _ = c.file.Write(data)
			}
			c.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

// WriteToFile writes content directly to the capture file without printing to terminal
// This is useful for writing verbose output that should only appear in the log file
func (c *Capture) WriteToFile(content string) {
	if content == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.file != nil {
		_, _ = c.file.WriteString(content)
	}
}

// Stop restores original stdout/stderr and closes the file
func (c *Capture) Stop() error {
	// Guard against nil receiver
	if c == nil {
		return nil
	}

	// Signal done to tee goroutines
	close(c.done)

	// Close pipe writers to signal EOF to tee goroutines
	if c.outWriter != nil {
		c.outWriter.Close()
	}
	if c.errWriter != nil {
		c.errWriter.Close()
	}

	// Restore original stdout/stderr immediately
	os.Stdout = c.originalOut
	os.Stderr = c.originalErr

	// Wait for tee goroutines to finish
	c.wg.Wait()

	// Close readers
	if c.outReader != nil {
		c.outReader.Close()
	}
	if c.errReader != nil {
		c.errReader.Close()
	}

	// Close file
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.file != nil {
		_ = c.file.Sync()
		c.file.Close()
		c.file = nil
	}

	return nil
}
