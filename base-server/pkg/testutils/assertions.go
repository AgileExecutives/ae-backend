package testutils
package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertTimeEqual asserts that two times are equal within a tolerance
func AssertTimeEqual(t *testing.T, expected, actual time.Time, tolerance time.Duration, msgAndArgs ...interface{}) {
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	
	assert.True(t, diff <= tolerance, 
		"Times differ by more than tolerance. Expected: %s, Actual: %s, Diff: %s, Tolerance: %s",
		expected, actual, diff, tolerance)
}

// AssertTimeNotZero asserts that a time is not zero
func AssertTimeNotZero(t *testing.T, actual time.Time, msgAndArgs ...interface{}) {
	assert.False(t, actual.IsZero(), msgAndArgs...)
}

// AssertTimeBefore asserts that time1 is before time2
func AssertTimeBefore(t *testing.T, time1, time2 time.Time, msgAndArgs ...interface{}) {
	assert.True(t, time1.Before(time2), 
		"Expected %s to be before %s", time1, time2)
}

// AssertTimeAfter asserts that time1 is after time2
func AssertTimeAfter(t *testing.T, time1, time2 time.Time, msgAndArgs ...interface{}) {
	assert.True(t, time1.After(time2), 
		"Expected %s to be after %s", time1, time2)
}

// RequireNoDBError fails the test if there's a database error
func RequireNoDBError(t *testing.T, err error, operation string) {
	require.NoError(t, err, "Database error during %s", operation)
}

// AssertFloatEqual asserts that two floats are equal within tolerance
func AssertFloatEqual(t *testing.T, expected, actual, tolerance float64, msgAndArgs ...interface{}) {
	diff := expected - actual
	if diff < 0 {
		diff = -diff
	}
	
	assert.True(t, diff <= tolerance,
		"Floats differ by more than tolerance. Expected: %f, Actual: %f, Diff: %f, Tolerance: %f",
		expected, actual, diff, tolerance)
}

// AssertDecimalEqual asserts decimal equality with 2 decimal places (for currency)
func AssertDecimalEqual(t *testing.T, expected, actual float64, msgAndArgs ...interface{}) {
	AssertFloatEqual(t, expected, actual, 0.01, msgAndArgs...)
}

// AssertPointerNotNil asserts that a pointer is not nil
func AssertPointerNotNil(t *testing.T, ptr interface{}, msgAndArgs ...interface{}) {
	assert.NotNil(t, ptr, msgAndArgs...)
}

// AssertPointerNil asserts that a pointer is nil
func AssertPointerNil(t *testing.T, ptr interface{}, msgAndArgs ...interface{}) {
	assert.Nil(t, ptr, msgAndArgs...)
}

// AssertSliceNotEmpty asserts that a slice is not empty
func AssertSliceNotEmpty(t *testing.T, slice interface{}, msgAndArgs ...interface{}) {
	assert.NotEmpty(t, slice, msgAndArgs...)
}

// AssertSliceLength asserts that a slice has expected length
func AssertSliceLength(t *testing.T, slice interface{}, expectedLen int, msgAndArgs ...interface{}) {
	assert.Len(t, slice, expectedLen, msgAndArgs...)
}

// AssertMapHasKey asserts that a map contains a key
func AssertMapHasKey(t *testing.T, m map[string]interface{}, key string, msgAndArgs ...interface{}) {
	_, ok := m[key]
	assert.True(t, ok, "Map should contain key: %s", key)
}

// AssertStringNotEmpty asserts that a string is not empty
func AssertStringNotEmpty(t *testing.T, str string, msgAndArgs ...interface{}) {
	assert.NotEmpty(t, str, msgAndArgs...)
}

// AssertUintNotZero asserts that a uint is not zero
func AssertUintNotZero(t *testing.T, val uint, msgAndArgs ...interface{}) {
	assert.NotZero(t, val, msgAndArgs...)
}

// AssertIDsMatch asserts that two IDs match
func AssertIDsMatch(t *testing.T, expected, actual uint, msgAndArgs ...interface{}) {
	assert.Equal(t, expected, actual, msgAndArgs...)
}
