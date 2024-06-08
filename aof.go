// An append-only file (AOF) is a type of data storage mechanism used in databases for ensuring
// durability of operations. In an AOF system, every operation performed on the database is written
// sequentially to the end of the file. This means that new data is always added to the file, but
// existing data is never modified or deleted. We use an AOF for our Redis-like database
// as a form of write-ahead logging (WAL). This provide a simple and efficient way to persist database
// operations, making data recovery easier in case of system failures or crashes. Since operations
// are only appended to the file, there's no risk of corruption due to simultaneous writes or data
// inconsistency. Additionally, AOF files are typically human-readable, making them easier to inspect
// and debug if necessary.

// AOF files typically log the operations themselves rather than the resulting values or the database
// contents after each operation. For example, if you set a key-value pair in the database, the AOF
// file would log the command to set that key-value pair rather than the actual value being set.
// Similarly, if you delete a key, the AOF file would log the delete command.
// This approach simplifies the logging process and reduces the amount of data that needs to be
// written to the AOF file, making it more efficient. Additionally, it allows for easier recovery
// and replication since the database can simply replay the operations stored in the AOF file to
// rebuild its state.
package main

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"
)

// creates a struct to manage an Aof file
type Aof struct {
	file *os.File
	rd   *bufio.Reader
	// ennsures one goroutine can write to file at a given time
	mu sync.Mutex
}

// NewAof is a function that creates and initializes a new Aof struct for managing an append-only file (AOF).
// It takes a file path as input and returns a pointer to the Aof struct and an error.
func NewAof(path string) (*Aof, error) {
	// Open or create a file at the specified path with read-write permissions (0666).
	// If the file does not exist, it will be created
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	// instance of Aof with os.file pointer f, and bufio.NewReader
	aof := &Aof{
		file: f,
		//rd wraps around f reading from it
		rd: bufio.NewReader(f),
	}

	// Start a goroutine to sync AOF to disk every 1 second
	go func() {
		for {
			aof.mu.Lock()

			aof.file.Sync()

			aof.mu.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (aof *Aof) Close() error {
	// Lock the mutex to ensure exclusive access to the AOF file during the close operation
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

func (aof *Aof) Write(value Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// Use defer to ensure that the mutex is unlocked after the write operation,
	// even if an error occurs. This guarantees that the mutex is always released
	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

// Read reads commands from the AOF file, parses them, and invokes the provided
// function for each command value. It ensures thread-safe access to the AOF file.
func (aof *Aof) Read(fn func(value Value)) error {
	// Lock the mutex to ensure exclusive access to the AOF file during the read operation.
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// Seek to the beginning of the AOF file to start reading from the start.
	aof.file.Seek(0, io.SeekStart)

	// Create a new rESP (Redis Serialization Protocol) reader for reading commands from the AOF file.
	reader := newrESP(aof.file)

	// Iterate over each command in the AOF file.
	for {
		// Read the next command value from the AOF file.
		value, err := reader.Read()
		if err != nil {
			// If an error occurs while reading:
			if err == io.EOF {
				// If the end of the file (EOF) is reached, break the loop.
				break
			}
			// Return the error if it's not EOF.
			return err
		}

		// Invoke the provided function with the command value.
		fn(value)
	}

	// Return nil to indicate that the read operation was successful.
	return nil
}
