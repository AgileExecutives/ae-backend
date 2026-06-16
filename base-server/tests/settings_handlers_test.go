package settings_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae/base-server/pkg/settings"
	"github.com/ae/base-server/pkg/settings/entities"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSettingsHandlers_SetAndGetOrganizationSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	system, err := settings.NewSettingsSystem(db)
	assert.NoError(t, err)

	router := gin.New()
	system.RegisterRoutes(router)

	// Set a setting
	body := entities.SettingRequest{
		Domain: "company",
		Key:    "locale",
		Data: map[string]interface{}{
			"value": "de-DE",
		},
	}
	payload, err := json.Marshal(body)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/organizations/123", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get all settings
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/settings/organizations/123", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var resp entities.SettingsResponse
	err = json.Unmarshal(w2.Body.Bytes(), &resp)
	assert.NoError(t, err)

	if assert.Contains(t, resp.Settings, "company") {
		assert.Equal(t, "de-DE", resp.Settings["company"]["locale"].(map[string]interface{})["value"])
	}
}
