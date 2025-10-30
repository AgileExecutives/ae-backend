package swagger

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	internalModules "github.com/ae-base-server/internal/modules"
	"github.com/swaggo/swag"
)

// ModuleSwaggerIntegration handles dynamic Swagger documentation generation
type ModuleSwaggerIntegration struct {
	baseDoc        *swag.Spec
	moduleRegistry *internalModules.DefaultModuleRegistry
}

// NewModuleSwaggerIntegration creates a new Swagger integration instance
func NewModuleSwaggerIntegration(registry *internalModules.DefaultModuleRegistry) *ModuleSwaggerIntegration {
	return &ModuleSwaggerIntegration{
		moduleRegistry: registry,
	}
}

// GetModuleRegistry returns the module registry instance
func (m *ModuleSwaggerIntegration) GetModuleRegistry() *internalModules.DefaultModuleRegistry {
	return m.moduleRegistry
}

// GenerateSwaggerDoc generates a complete Swagger document with module endpoints
func (m *ModuleSwaggerIntegration) GenerateSwaggerDoc() (map[string]interface{}, error) {
	// Get the base Swagger spec
	baseSpec := swag.GetSwagger("swagger")
	if baseSpec == nil {
		return nil, fmt.Errorf("base swagger spec not found")
	}

	// Parse the base spec
	var baseDoc map[string]interface{}
	if err := json.Unmarshal([]byte(baseSpec.ReadDoc()), &baseDoc); err != nil {
		return nil, fmt.Errorf("failed to parse base swagger spec: %w", err)
	}

	// Add module information to the spec
	if err := m.addModulesToSwaggerDoc(baseDoc); err != nil {
		return nil, fmt.Errorf("failed to add modules to swagger doc: %w", err)
	}

	return baseDoc, nil
}

// addModulesToSwaggerDoc adds module endpoints and tags to the base Swagger document
func (m *ModuleSwaggerIntegration) addModulesToSwaggerDoc(baseDoc map[string]interface{}) error {
	// Get module information
	moduleInfos := m.GetModuleSwaggerInfo()

	// Add module tags to the main document
	if err := m.addModuleTags(baseDoc, moduleInfos); err != nil {
		return err
	}

	// Add module paths to the main document
	if err := m.addModulePaths(baseDoc, moduleInfos); err != nil {
		return err
	}

	// Add module definitions/schemas
	if err := m.addModuleDefinitions(baseDoc, moduleInfos); err != nil {
		return err
	}

	log.Printf("📋 Successfully integrated %d modules into Swagger documentation", len(moduleInfos))
	return nil
}

// ModuleSwaggerInfo represents Swagger information for a module
type ModuleSwaggerInfo struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Version     string             `json:"version"`
	Tags        []ModuleSwaggerTag `json:"tags"`
}

// ModuleSwaggerTag represents a Swagger tag for a module
type ModuleSwaggerTag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetModuleSwaggerInfo extracts Swagger information from registered modules
func (m *ModuleSwaggerIntegration) GetModuleSwaggerInfo() map[string]ModuleSwaggerInfo {
	moduleInfos := make(map[string]ModuleSwaggerInfo)

	// Get all enabled modules
	enabledModules := m.moduleRegistry.GetEnabledModules()

	for _, module := range enabledModules {
		name := module.GetName()

		// Generate tags based on module name and common patterns
		var tags []ModuleSwaggerTag

		switch name {
		case "calendar":
			tags = []ModuleSwaggerTag{
				{Name: "calendars", Description: "Calendar management operations"},
				{Name: "recurring-events", Description: "Recurring event management operations"},
				{Name: "events", Description: "Event management operations"},
			}
		case "demo":
			tags = []ModuleSwaggerTag{
				{Name: "demo", Description: "Demo module operations"},
			}
		default:
			// Generic tags for unknown modules
			tags = []ModuleSwaggerTag{
				{Name: name, Description: fmt.Sprintf("%s operations", strings.Title(name))},
			}
		}

		// Create basic Swagger info from module metadata
		swaggerInfo := ModuleSwaggerInfo{
			Title:       strings.Title(name) + " Module",
			Description: fmt.Sprintf("API endpoints for the %s module", name),
			Version:     module.GetVersion(),
			Tags:        tags,
		}

		moduleInfos[name] = swaggerInfo
	}

	return moduleInfos
}

// addModuleTags adds tags from all modules to the base Swagger document
func (m *ModuleSwaggerIntegration) addModuleTags(baseDoc map[string]interface{}, moduleInfos map[string]ModuleSwaggerInfo) error {
	tags, exists := baseDoc["tags"]
	if !exists {
		baseDoc["tags"] = make([]interface{}, 0)
		tags = baseDoc["tags"]
	}

	tagsList, ok := tags.([]interface{})
	if !ok {
		tagsList = make([]interface{}, 0)
	}

	// Collect all module tags
	for moduleName, moduleInfo := range moduleInfos {
		for _, tag := range moduleInfo.Tags {
			// Prefix module tags to avoid conflicts
			moduleTag := map[string]interface{}{
				"name":        fmt.Sprintf("%s-%s", moduleName, tag.Name),
				"description": fmt.Sprintf("[%s] %s", strings.Title(moduleName), tag.Description),
			}
			tagsList = append(tagsList, moduleTag)
		}
	}

	baseDoc["tags"] = tagsList
	return nil
}

// addModulePaths adds module endpoints to the base Swagger document
func (m *ModuleSwaggerIntegration) addModulePaths(baseDoc map[string]interface{}, moduleInfos map[string]ModuleSwaggerInfo) error {
	paths, exists := baseDoc["paths"]
	if !exists {
		baseDoc["paths"] = make(map[string]interface{})
		paths = baseDoc["paths"]
	}

	pathsMap, ok := paths.(map[string]interface{})
	if !ok {
		pathsMap = make(map[string]interface{})
	}

	// For now, we'll add placeholder paths for each module
	// In a full implementation, you'd scan the module handlers for actual paths
	for moduleName, moduleInfo := range moduleInfos {
		basePath := fmt.Sprintf("/api/v1/modules/%s", moduleName)

		// Add basic CRUD paths for each tag (representing entities)
		for _, tag := range moduleInfo.Tags {
			entityPath := fmt.Sprintf("%s/%s", basePath, tag.Name)

			// Add GET collection endpoint
			pathsMap[entityPath] = map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{fmt.Sprintf("%s-%s", moduleName, tag.Name)},
					"summary":     fmt.Sprintf("Get all %s", tag.Name),
					"description": fmt.Sprintf("Retrieve all %s from %s module", tag.Name, moduleName),
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Success",
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"data": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"type": "object",
										},
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
				"post": map[string]interface{}{
					"tags":        []string{fmt.Sprintf("%s-%s", moduleName, tag.Name)},
					"summary":     fmt.Sprintf("Create %s", strings.TrimSuffix(tag.Name, "s")),
					"description": fmt.Sprintf("Create a new %s in %s module", strings.TrimSuffix(tag.Name, "s"), moduleName),
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"consumes":    []string{"application/json"},
					"produces":    []string{"application/json"},
					"parameters": []map[string]interface{}{
						{
							"name":        "body",
							"in":          "body",
							"required":    true,
							"description": fmt.Sprintf("%s data", strings.Title(strings.TrimSuffix(tag.Name, "s"))),
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Created",
						},
						"400": map[string]interface{}{
							"description": "Bad Request",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			}

			// Add GET by ID endpoint
			itemPath := fmt.Sprintf("%s/{id}", entityPath)
			pathsMap[itemPath] = map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{fmt.Sprintf("%s-%s", moduleName, tag.Name)},
					"summary":     fmt.Sprintf("Get %s by ID", strings.TrimSuffix(tag.Name, "s")),
					"description": fmt.Sprintf("Retrieve a specific %s by ID from %s module", strings.TrimSuffix(tag.Name, "s"), moduleName),
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"type":        "integer",
							"description": fmt.Sprintf("%s ID", strings.Title(strings.TrimSuffix(tag.Name, "s"))),
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Success",
						},
						"404": map[string]interface{}{
							"description": "Not Found",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
				"put": map[string]interface{}{
					"tags":        []string{fmt.Sprintf("%s-%s", moduleName, tag.Name)},
					"summary":     fmt.Sprintf("Update %s", strings.TrimSuffix(tag.Name, "s")),
					"description": fmt.Sprintf("Update an existing %s in %s module", strings.TrimSuffix(tag.Name, "s"), moduleName),
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"consumes":    []string{"application/json"},
					"produces":    []string{"application/json"},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"type":        "integer",
							"description": fmt.Sprintf("%s ID", strings.Title(strings.TrimSuffix(tag.Name, "s"))),
						},
						{
							"name":        "body",
							"in":          "body",
							"required":    true,
							"description": fmt.Sprintf("Updated %s data", strings.TrimSuffix(tag.Name, "s")),
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Updated",
						},
						"400": map[string]interface{}{
							"description": "Bad Request",
						},
						"404": map[string]interface{}{
							"description": "Not Found",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
				"delete": map[string]interface{}{
					"tags":        []string{fmt.Sprintf("%s-%s", moduleName, tag.Name)},
					"summary":     fmt.Sprintf("Delete %s", strings.TrimSuffix(tag.Name, "s")),
					"description": fmt.Sprintf("Delete a %s from %s module", strings.TrimSuffix(tag.Name, "s"), moduleName),
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"type":        "integer",
							"description": fmt.Sprintf("%s ID", strings.Title(strings.TrimSuffix(tag.Name, "s"))),
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Deleted",
						},
						"404": map[string]interface{}{
							"description": "Not Found",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			}
		}
	}

	baseDoc["paths"] = pathsMap
	return nil
}

// addModuleDefinitions adds module schema definitions to the base Swagger document
func (m *ModuleSwaggerIntegration) addModuleDefinitions(baseDoc map[string]interface{}, moduleInfos map[string]ModuleSwaggerInfo) error {
	definitions, exists := baseDoc["definitions"]
	if !exists {
		baseDoc["definitions"] = make(map[string]interface{})
		definitions = baseDoc["definitions"]
	}

	definitionsMap, ok := definitions.(map[string]interface{})
	if !ok {
		definitionsMap = make(map[string]interface{})
	}

	// Add basic schemas for each module
	for moduleName, moduleInfo := range moduleInfos {
		// Add module info schema
		definitionsMap[fmt.Sprintf("%sModuleInfo", strings.Title(moduleName))] = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title":       map[string]interface{}{"type": "string"},
				"description": map[string]interface{}{"type": "string"},
				"version":     map[string]interface{}{"type": "string"},
			},
		}

		// Add basic schemas for each tag/entity
		for _, tag := range moduleInfo.Tags {
			entityName := strings.Title(strings.TrimSuffix(tag.Name, "s"))
			definitionsMap[fmt.Sprintf("%s%s", strings.Title(moduleName), entityName)] = map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":         map[string]interface{}{"type": "integer"},
					"created_at": map[string]interface{}{"type": "string", "format": "date-time"},
					"updated_at": map[string]interface{}{"type": "string", "format": "date-time"},
				},
			}
		}
	}

	baseDoc["definitions"] = definitionsMap
	return nil
}
