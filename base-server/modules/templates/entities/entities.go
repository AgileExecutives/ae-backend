package entities

import (
	"encoding/json"
)

// MarshalJSON is a helper to marshal data to JSON
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON is a helper to unmarshal JSON data
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
