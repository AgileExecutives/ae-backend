package renderer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ContractProvider interface for getting contract schemas
type ContractProvider interface {
	GetContract(templateKey string) (map[string]interface{}, error)
	GetAvailableContracts() ([]string, error)
}

// DBRenderer handles template rendering with database-stored contracts
type DBRenderer struct {
	templatesDir     string
	contractProvider ContractProvider
	schemas          map[string]*jsonschema.Schema
}

// NewDBRenderer creates a new renderer with database-stored contracts
func NewDBRenderer(templatesDir string, contractProvider ContractProvider) (*DBRenderer, error) {
	r := &DBRenderer{
		templatesDir:     templatesDir,
		contractProvider: contractProvider,
		schemas:          make(map[string]*jsonschema.Schema),
	}

	// Load all schema contracts from database
	if err := r.loadSchemasFromDB(); err != nil {
		return nil, fmt.Errorf("failed to load schemas from database: %w", err)
	}

	return r, nil
}

// loadSchemasFromDB loads all schema contracts from the database
func (r *DBRenderer) loadSchemasFromDB() error {
	contracts, err := r.contractProvider.GetAvailableContracts()
	if err != nil {
		return err
	}

	for _, contractName := range contracts {
		schemaData, err := r.contractProvider.GetContract(contractName)
		if err != nil {
			return fmt.Errorf("failed to get contract %s: %w", contractName, err)
		}

		schemaBytes, err := json.Marshal(schemaData)
		if err != nil {
			return fmt.Errorf("failed to marshal schema %s: %w", contractName, err)
		}

		// Compile schema using jsonschema library
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(fmt.Sprintf("schema_%s.json", contractName), strings.NewReader(string(schemaBytes))); err == nil {
			if compiledSchema, err := compiler.Compile(fmt.Sprintf("schema_%s.json", contractName)); err == nil {
				r.schemas[contractName] = compiledSchema
			}
		}
	}

	return nil
}

// RenderTemplate renders a template with contract validation using database contracts
func (r *DBRenderer) RenderTemplate(templateName, contractName string, jsonData []byte) ([]byte, error) {
	// Step 1: Validate against schema contract from database
	if err := r.validateContract(contractName, jsonData); err != nil {
		return nil, fmt.Errorf("contract validation failed: %w", err)
	}

	// Step 2: Unmarshal to map[string]any (no domain types!)
	var data map[string]any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Step 3: Load and parse template
	templatePath := filepath.Join(r.templatesDir, templateName+".html")
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New(templateName).Option("missingkey=zero").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Step 4: Render template (pure data-driven)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// RenderTemplateFromContent renders a template from content string with contract validation
func (r *DBRenderer) RenderTemplateFromContent(templateContent, contractName string, jsonData []byte) ([]byte, error) {
	// Step 1: Validate against schema contract from database
	if err := r.validateContract(contractName, jsonData); err != nil {
		return nil, fmt.Errorf("contract validation failed: %w", err)
	}

	// Step 2: Unmarshal to map[string]any (no domain types!)
	var data map[string]any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Step 3: Parse template from content
	tmpl, err := template.New("template").Option("missingkey=zero").Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Step 4: Render template (pure data-driven)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// validateContract validates JSON data against the specified schema contract from database
func (r *DBRenderer) validateContract(contractName string, jsonData []byte) error {
	schema, exists := r.schemas[contractName]
	if !exists {
		return fmt.Errorf("unknown contract: %s", contractName)
	}

	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if err := schema.Validate(data); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// GetAvailableContracts returns list of available schema contracts from database
func (r *DBRenderer) GetAvailableContracts() []string {
	contracts := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		contracts = append(contracts, name)
	}
	return contracts
}
