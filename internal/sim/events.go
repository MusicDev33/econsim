package sim

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Event represents a simulation event in NDJSON format
type Event struct {
	Tick int         `json:"tick"`
	Type string      `json:"type"`
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

// EventEmitter writes events to an NDJSON file
type EventEmitter struct {
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

// NewEventEmitter creates a new event emitter that writes to the given file path.
// Creates the parent directory if it doesn't exist. Truncates any existing file.
func NewEventEmitter(path string) (*EventEmitter, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Create (or truncate) the file
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &EventEmitter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Emit writes an event to the NDJSON file
func (e *EventEmitter) Emit(tick int, name string, data interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	event := Event{
		Tick: tick,
		Type: "event",
		Name: name,
		Data: data,
	}

	return e.encoder.Encode(event)
}

// Close closes the underlying file
func (e *EventEmitter) Close() error {
	return e.file.Close()
}
