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

// Renderer handles template rendering with JSON contract validation
type Renderer struct {
	templatesDir string
	contractsDir string
	schemas      map[string]*jsonschema.Schema
}

// New creates a new Renderer instance
func New(templatesDir, contractsDir string) (*Renderer, error) {
	r := &Renderer{
		templatesDir: templatesDir,
		contractsDir: contractsDir,
		schemas:      make(map[string]*jsonschema.Schema),
	}

	// Load all schema contracts
	if err := r.loadSchemas(); err != nil {
		return nil, fmt.Errorf("failed to load schemas: %w", err)
	}

	return r, nil
}

// loadSchemas loads all JSON schema files from the contracts directory
func (r *Renderer) loadSchemas() error {
	files, err := filepath.Glob(filepath.Join(r.contractsDir, "*.json"))
	if err != nil {
		return err
	}

	compiler := jsonschema.NewCompiler()

	for _, file := range files {
		schemaBytes, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read schema file %s: %w", file, err)
		}

		// Extract schema name from filename (e.g., "invoice-contract.json" -> "invoice")
		fileName := filepath.Base(file)
		var schemaName string
		if filepath.Ext(fileName) == ".json" {
			baseName := fileName[:len(fileName)-len(".json")]
			if strings.HasSuffix(baseName, "-contract") {
				schemaName = baseName[:len(baseName)-len("-contract")]
			} else {
				schemaName = baseName
			}
		}

		schemaURL := fmt.Sprintf("file://%s", file)
		if err := compiler.AddResource(schemaURL, bytes.NewReader(schemaBytes)); err != nil {
			return fmt.Errorf("failed to add schema resource %s: %w", file, err)
		}

		schema, err := compiler.Compile(schemaURL)
		if err != nil {
			return fmt.Errorf("failed to compile schema %s: %w", file, err)
		}

		r.schemas[schemaName] = schema
	}

	return nil
}

// RenderTemplate renders a template with data validation against contract
func (r *Renderer) RenderTemplate(templateName, contractName string, jsonData []byte) ([]byte, error) {
	// Step 1: Validate against schema contract
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

// RenderTemplateFromContent renders a template from content string with data validation
func (r *Renderer) RenderTemplateFromContent(templateContent, contractName string, jsonData []byte) ([]byte, error) {
	// Step 1: Validate against schema contract
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

// validateContract validates JSON data against the specified schema contract
func (r *Renderer) validateContract(contractName string, jsonData []byte) error {
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

// GetAvailableContracts returns list of available schema contracts
func (r *Renderer) GetAvailableContracts() []string {
	contracts := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		contracts = append(contracts, name)
	}
	return contracts
}
