package utils

import (
	"time"
)

// ParseTime parses a time string in RFC3339 format
func ParseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, timeStr)
}
