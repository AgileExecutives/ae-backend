package entities

import (
	"encoding/json"
)

// Helper functions for JSON marshaling/unmarshaling

// MarshalJSON marshals data to JSON bytes
func MarshalJSON(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// UnmarshalJSON unmarshals JSON bytes to data
func UnmarshalJSON(jsonData []byte, v interface{}) error {
	return json.Unmarshal(jsonData, v)
}
