package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// SetupTestRouter creates a test router with Gin in test mode
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add recovery middleware for tests
	router.Use(gin.Recovery())

	return router
}

// SetupAuthContext sets authentication context in Gin context
func SetupAuthContext(c *gin.Context, tenantID, userID uint) {
	c.Set("tenant_id", tenantID)
	c.Set("user_id", userID)
}

// SetupAuthContextWithOrg sets authentication context with organization
func SetupAuthContextWithOrg(c *gin.Context, tenantID, userID, orgID uint) {
	c.Set("tenant_id", tenantID)
	c.Set("user_id", userID)
	c.Set("organization_id", orgID)
}

// MakeJSONRequest creates and executes a JSON HTTP request
func MakeJSONRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, path, bodyReader)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// MakeAuthenticatedRequest creates an authenticated HTTP request
func MakeAuthenticatedRequest(router *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, path, bodyReader)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// ParseJSONResponse parses JSON response body into target struct
func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	require.NoError(t, err, "Failed to parse JSON response")
}

// AssertJSONResponse asserts status code and parses response
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	require.Equal(t, expectedStatus, w.Code, "Unexpected status code. Response: %s", w.Body.String())

	if target != nil && w.Code < 300 {
		ParseJSONResponse(t, w, target)
	}
}

// AssertErrorResponse asserts error response structure
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessageContains string) {
	require.Equal(t, expectedStatus, w.Code, "Unexpected status code")

	var errorResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err, "Failed to parse error response")

	if expectedMessageContains != "" {
		message, ok := errorResp["error"].(string)
		if !ok {
			message, _ = errorResp["message"].(string)
		}

		require.Contains(t, message, expectedMessageContains, "Error message doesn't contain expected text")
	}
}

// CreateTestContext creates a Gin context for testing handlers
func CreateTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	return c, w
}

// CreateAuthenticatedTestContext creates an authenticated Gin context
func CreateAuthenticatedTestContext(tenantID, userID uint) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := CreateTestContext()
	SetupAuthContext(c, tenantID, userID)

	return c, w
}

// SetJSONBody sets a JSON body in the test context
func SetJSONBody(c *gin.Context, body interface{}) {
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
}

// SetURLParam sets a URL parameter in the context
func SetURLParam(c *gin.Context, key, value string) {
	c.Params = append(c.Params, gin.Param{Key: key, Value: value})
}

// SetQueryParam sets a query parameter in the context
func SetQueryParam(c *gin.Context, key, value string) {
	if c.Request.URL.RawQuery == "" {
		c.Request.URL.RawQuery = key + "=" + value
	} else {
		c.Request.URL.RawQuery += "&" + key + "=" + value
	}
}
