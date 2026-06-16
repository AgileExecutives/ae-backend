package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	baseAPI "github.com/ae/base-server/api"
	"github.com/ae/base-server/pkg/core"
	settingsEntities "github.com/ae/base-server/pkg/settings/entities"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientEntities "github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/handlers"
	clientServices "github.com/unburdy/unburdy-server-api/modules/client_management/services"
	cmsettings "github.com/unburdy/unburdy-server-api/modules/client_management/settings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRouteProvider_RegistersOrganizationSettingsCompatibilityRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_foreign_keys=on"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&settingsEntities.SettingDefinition{}, &settingsEntities.Setting{}, &clientEntities.RegistrationToken{}))

	for _, def := range cmsettings.GetBillingSettingsDefinitions() {
		payload, marshalErr := json.Marshal(def.Data)
		require.NoError(t, marshalErr)
		require.NoError(t, db.Create(&settingsEntities.SettingDefinition{
			Domain:  def.Domain,
			Key:     def.Key,
			Version: def.Version,
			Data:    payload,
		}).Error)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", &baseAPI.User{ID: 1, TenantID: 1, OrganizationID: 1})
		c.Next()
	})

	clientService := clientServices.NewClientService(db)
	provider := NewRouteProvider(nil, nil, nil, nil, nil, nil, nil, handlers.NewOrganizationSettingsHandler(db, clientService), db)
	provider.RegisterRoutes(router.Group("/api/v1"), &core.ModuleContext{Router: router})

	t.Run("registration settings route exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organization/settings/registration", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), "registration")
	})

	t.Run("registration settings can be updated", func(t *testing.T) {
		payload := []byte(`{"email_verification_enabled":false,"required_fields":["first_name","email"]}`)
		putReq := httptest.NewRequest(http.MethodPut, "/api/v1/organization/settings/registration", bytes.NewReader(payload))
		putReq.Header.Set("Content-Type", "application/json")
		putRes := httptest.NewRecorder()

		router.ServeHTTP(putRes, putReq)

		require.Equal(t, http.StatusOK, putRes.Code)
		assert.Contains(t, putRes.Body.String(), "Registration settings updated successfully")

		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/organization/settings/registration", nil)
		getRes := httptest.NewRecorder()

		router.ServeHTTP(getRes, getReq)

		require.Equal(t, http.StatusOK, getRes.Code)
		assert.Contains(t, getRes.Body.String(), `"email_verification_enabled":false`)
		assert.Contains(t, getRes.Body.String(), `"required_fields":["first_name","email"]`)
	})

	t.Run("registration settings support partial updates", func(t *testing.T) {
		initialPayload := []byte(`{"registration_headline":"Custom headline","registration_intro_text":"Custom intro"}`)
		initialReq := httptest.NewRequest(http.MethodPut, "/api/v1/organization/settings/registration", bytes.NewReader(initialPayload))
		initialReq.Header.Set("Content-Type", "application/json")
		initialRes := httptest.NewRecorder()

		router.ServeHTTP(initialRes, initialReq)

		require.Equal(t, http.StatusOK, initialRes.Code)

		partialPayload := []byte(`{"email_verification_enabled":false}`)
		partialReq := httptest.NewRequest(http.MethodPut, "/api/v1/organization/settings/registration", bytes.NewReader(partialPayload))
		partialReq.Header.Set("Content-Type", "application/json")
		partialRes := httptest.NewRecorder()

		router.ServeHTTP(partialRes, partialReq)

		require.Equal(t, http.StatusOK, partialRes.Code)

		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/organization/settings/registration", nil)
		getRes := httptest.NewRecorder()

		router.ServeHTTP(getRes, getReq)

		require.Equal(t, http.StatusOK, getRes.Code)
		assert.Contains(t, getRes.Body.String(), `"registration_headline":"Custom headline"`)
		assert.Contains(t, getRes.Body.String(), `"registration_intro_text":"Custom intro"`)
		assert.Contains(t, getRes.Body.String(), `"email_verification_enabled":false`)
	})

	t.Run("registration settings accept wrapped payloads", func(t *testing.T) {
		payload := []byte(`{"settings":{"registration_headline":"Wrapped headline","email_verification_enabled":false}}`)
		putReq := httptest.NewRequest(http.MethodPut, "/api/v1/organization/settings/registration", bytes.NewReader(payload))
		putReq.Header.Set("Content-Type", "application/json")
		putRes := httptest.NewRecorder()

		router.ServeHTTP(putRes, putReq)

		require.Equal(t, http.StatusOK, putRes.Code)

		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/organization/settings/registration", nil)
		getRes := httptest.NewRecorder()

		router.ServeHTTP(getRes, getReq)

		require.Equal(t, http.StatusOK, getRes.Code)
		assert.Contains(t, getRes.Body.String(), `"registration_headline":"Wrapped headline"`)
		assert.Contains(t, getRes.Body.String(), `"email_verification_enabled":false`)
	})

	t.Run("registration settings are available via registration token", func(t *testing.T) {
		require.NoError(t, db.Create(&clientEntities.RegistrationToken{
			TenantID:       1,
			OrganizationID: 1,
			Token:          "public-token",
			CreatedBy:      1,
		}).Error)

		payload := []byte(`{"registration_headline":"Public headline","registration_intro_text":"Public intro"}`)
		putReq := httptest.NewRequest(http.MethodPut, "/api/v1/organization/settings/registration", bytes.NewReader(payload))
		putReq.Header.Set("Content-Type", "application/json")
		putRes := httptest.NewRecorder()

		router.ServeHTTP(putRes, putReq)

		require.Equal(t, http.StatusOK, putRes.Code)

		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/clients/registration-settings/public-token", nil)
		getRes := httptest.NewRecorder()

		router.ServeHTTP(getRes, getReq)

		require.Equal(t, http.StatusOK, getRes.Code)
		assert.Contains(t, getRes.Body.String(), `"registration_headline":"Public headline"`)
		assert.Contains(t, getRes.Body.String(), `"registration_intro_text":"Public intro"`)
	})

	t.Run("billing settings route exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organization/settings/billing?organization_id=1", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), cmsettings.DomainBilling)
		assert.Contains(t, res.Body.String(), "payment_due_days")
	})

	t.Run("billing settings validates organization_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organization/settings/billing?organization_id=bad", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		require.Equal(t, http.StatusBadRequest, res.Code)
	})
}