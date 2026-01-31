package testutils
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
func MakeJSONRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)






































































































}	}		c.Request.URL.RawQuery += "&" + key + "=" + value	} else {		c.Request.URL.RawQuery = key + "=" + value	if c.Request.URL.RawQuery == "" {func SetQueryParam(c *gin.Context, key, value string) {// SetQueryParam sets a query parameter in the context}	c.Params = append(c.Params, gin.Param{Key: key, Value: value})func SetURLParam(c *gin.Context, key, value string) {// SetURLParam sets a URL parameter in the context}	c.Request.Header.Set("Content-Type", "application/json")	c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))	jsonBody, _ := json.Marshal(body)func SetJSONBody(c *gin.Context, body interface{}) {// SetJSONBody sets a JSON body in the test context}	return c, w	SetupAuthContext(c, tenantID, userID)	c, w := CreateTestContext()func CreateAuthenticatedTestContext(tenantID, userID uint) (*gin.Context, *httptest.ResponseRecorder) {// CreateAuthenticatedTestContext creates an authenticated Gin context}	return c, w	c, _ := gin.CreateTestContext(w)	w := httptest.NewRecorder()func CreateTestContext() (*gin.Context, *httptest.ResponseRecorder) {// CreateTestContext creates a Gin context for testing handlers}	}		require.Contains(t, message, expectedMessageContains, "Error message doesn't contain expected text")		}			message, _ = errorResp["message"].(string)		if !ok {		message, ok := errorResp["error"].(string)	if expectedMessageContains != "" {		require.NoError(t, err, "Failed to parse error response")	err := json.Unmarshal(w.Body.Bytes(), &errorResp)	var errorResp map[string]interface{}		require.Equal(t, expectedStatus, w.Code, "Unexpected status code")func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessageContains string) {// AssertErrorResponse asserts error response structure}	}		ParseJSONResponse(t, w, target)	if target != nil && w.Code < 300 {		require.Equal(t, expectedStatus, w.Code, "Unexpected status code. Response: %s", w.Body.String())func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, target interface{}) {// AssertJSONResponse asserts status code and parses response}	require.NoError(t, err, "Failed to parse JSON response")	err := json.Unmarshal(w.Body.Bytes(), target)func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {// ParseJSONResponse parses JSON response body into target struct}	return w		router.ServeHTTP(w, req)	w := httptest.NewRecorder()		}		req.Header.Set("Authorization", "Bearer "+token)	if token != "" {	}		req.Header.Set("Content-Type", "application/json")	if body != nil {	req, _ := http.NewRequest(method, path, bodyReader)		}		bodyReader = bytes.NewBuffer(jsonBody)		jsonBody, _ := json.Marshal(body)	if body != nil {		var bodyReader io.Readerfunc MakeAuthenticatedRequest(router *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {// MakeAuthenticatedRequest creates an authenticated HTTP request}	return w		router.ServeHTTP(w, req)	w := httptest.NewRecorder()		}		req.Header.Set("Content-Type", "application/json")	if body != nil {	req, _ := http.NewRequest(method, path, bodyReader)		}