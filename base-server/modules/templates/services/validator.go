package services

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidationError represents a validation error with field path
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// TemplateValidator validates template data against schemas
type TemplateValidator struct{}

// NewTemplateValidator creates a new template validator
func NewTemplateValidator() *TemplateValidator {
	return &TemplateValidator{}
}

// Validate validates data against a JSON schema
// Schema format:
//
//	{
//	  "field_name": {
//	    "type": "string|number|boolean|object|array",
//	    "required": true|false,
//	    "properties": {...} // for nested objects
//	  }
//	}
func (v *TemplateValidator) Validate(schema map[string]interface{}, data map[string]interface{}) error {
	var errors ValidationErrors

	// Validate all schema fields
	for fieldName, fieldSchema := range schema {
		schemaMap, ok := fieldSchema.(map[string]interface{})
		if !ok {
			continue
		}

		// Check required fields
		required, _ := schemaMap["required"].(bool)
		value, exists := data[fieldName]

		if required && !exists {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: "field is required",
			})
			continue
		}

		if !exists {
			continue
		}

		// Validate type
		expectedType, _ := schemaMap["type"].(string)
		if expectedType != "" {
			if err := v.validateType(fieldName, expectedType, value, schemaMap); err != nil {
				errors = append(errors, err...)
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateType validates that a value matches the expected type
func (v *TemplateValidator) validateType(fieldName, expectedType string, value interface{}, schema map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("expected string, got %s", reflect.TypeOf(value).String()),
			})
		}

	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid number type
		default:
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("expected number, got %s", reflect.TypeOf(value).String()),
			})
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("expected boolean, got %s", reflect.TypeOf(value).String()),
			})
		}

	case "object":
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("expected object, got %s", reflect.TypeOf(value).String()),
			})
			return errors
		}

		// Validate nested properties
		if properties, ok := schema["properties"].(map[string]interface{}); ok {
			for propName, propSchema := range properties {
				propSchemaMap, ok := propSchema.(map[string]interface{})
				if !ok {
					continue
				}

				propValue, propExists := valueMap[propName]
				propRequired, _ := propSchemaMap["required"].(bool)

				if propRequired && !propExists {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("%s.%s", fieldName, propName),
						Message: "field is required",
					})
					continue
				}

				if !propExists {
					continue
				}

				propType, _ := propSchemaMap["type"].(string)
				if propType != "" {
					if err := v.validateType(fmt.Sprintf("%s.%s", fieldName, propName), propType, propValue, propSchemaMap); err != nil {
						errors = append(errors, err...)
					}
				}
			}
		}

	case "array":
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("expected array, got %s", reflect.TypeOf(value).String()),
			})
			return errors
		}

		// Validate array items if schema specified
		if itemSchema, ok := schema["items"].(map[string]interface{}); ok {
			items := reflect.ValueOf(value)
			for i := 0; i < items.Len(); i++ {
				item := items.Index(i).Interface()
				itemType, _ := itemSchema["type"].(string)
				if itemType != "" {
					if err := v.validateType(fmt.Sprintf("%s[%d]", fieldName, i), itemType, item, itemSchema); err != nil {
						errors = append(errors, err...)
					}
				}
			}
		}

	default:
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("unknown type: %s", expectedType),
		})
	}

	return errors
}

// ValidateWithSampleData validates that sample data matches the schema
func (v *TemplateValidator) ValidateWithSampleData(schema, sampleData map[string]interface{}) error {
	return v.Validate(schema, sampleData)
}

// ExtractRequiredFields returns a list of required field paths from a schema
func (v *TemplateValidator) ExtractRequiredFields(schema map[string]interface{}) []string {
	var required []string
	v.extractRequiredFieldsRecursive(schema, "", &required)
	return required
}

// extractRequiredFieldsRecursive recursively extracts required field paths
func (v *TemplateValidator) extractRequiredFieldsRecursive(schema map[string]interface{}, prefix string, required *[]string) {
	for fieldName, fieldSchema := range schema {
		schemaMap, ok := fieldSchema.(map[string]interface{})
		if !ok {
			continue
		}

		fieldPath := fieldName
		if prefix != "" {
			fieldPath = prefix + "." + fieldName
		}

		if isRequired, _ := schemaMap["required"].(bool); isRequired {
			*required = append(*required, fieldPath)
		}

		// Recursively check nested properties
		if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
			v.extractRequiredFieldsRecursive(properties, fieldPath, required)
		}
	}
}
