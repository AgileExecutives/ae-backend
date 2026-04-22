package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{Field: "name", Message: "field is required"}
	assert.Equal(t, "name: field is required", err.Error())
}

func TestValidationErrors_Error(t *testing.T) {
	t.Run("multiple errors joined", func(t *testing.T) {
		errs := ValidationErrors{
			{Field: "name", Message: "required"},
			{Field: "age", Message: "must be number"},
		}
		msg := errs.Error()
		assert.Contains(t, msg, "name: required")
		assert.Contains(t, msg, "age: must be number")
	})

	t.Run("empty errors", func(t *testing.T) {
		errs := ValidationErrors{}
		assert.Equal(t, "", errs.Error())
	})
}

func TestNewTemplateValidator(t *testing.T) {
	v := NewTemplateValidator()
	require.NotNil(t, v)
}

func TestTemplateValidator_Validate(t *testing.T) {
	v := NewTemplateValidator()

	schema := map[string]interface{}{
		"name": map[string]interface{}{"type": "string", "required": true},
		"age":  map[string]interface{}{"type": "number", "required": false},
	}

	t.Run("valid data passes", func(t *testing.T) {
		data := map[string]interface{}{"name": "Alice", "age": 30}
		err := v.Validate(schema, data)
		assert.NoError(t, err)
	})

	t.Run("missing required field fails", func(t *testing.T) {
		data := map[string]interface{}{"age": 30}
		err := v.Validate(schema, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("wrong type fails", func(t *testing.T) {
		data := map[string]interface{}{"name": 123}
		err := v.Validate(schema, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("optional field absent is fine", func(t *testing.T) {
		data := map[string]interface{}{"name": "Bob"}
		err := v.Validate(schema, data)
		assert.NoError(t, err)
	})

	t.Run("field schema not a map is skipped", func(t *testing.T) {
		badSchema := map[string]interface{}{
			"name": "not-a-map",
		}
		err := v.Validate(badSchema, map[string]interface{}{"name": "test"})
		assert.NoError(t, err)
	})
}

func TestTemplateValidator_ValidateType_String(t *testing.T) {
	v := NewTemplateValidator()
	schema := map[string]interface{}{
		"title": map[string]interface{}{"type": "string"},
	}

	t.Run("valid string", func(t *testing.T) {
		assert.NoError(t, v.Validate(schema, map[string]interface{}{"title": "hello"}))
	})

	t.Run("invalid string (number instead)", func(t *testing.T) {
		err := v.Validate(schema, map[string]interface{}{"title": 42})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})
}

func TestTemplateValidator_ValidateType_Number(t *testing.T) {
	v := NewTemplateValidator()
	schema := map[string]interface{}{
		"count": map[string]interface{}{"type": "number"},
	}

	for _, valid := range []interface{}{int(1), int8(1), int16(1), int32(1), int64(1),
		uint(1), uint8(1), uint16(1), uint32(1), uint64(1), float32(1.0), float64(1.0)} {
		assert.NoError(t, v.Validate(schema, map[string]interface{}{"count": valid}))
	}

	t.Run("string is invalid number", func(t *testing.T) {
		err := v.Validate(schema, map[string]interface{}{"count": "not-a-number"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "count")
	})
}

func TestTemplateValidator_ValidateType_Boolean(t *testing.T) {
	v := NewTemplateValidator()
	schema := map[string]interface{}{
		"active": map[string]interface{}{"type": "boolean"},
	}

	t.Run("valid bool true", func(t *testing.T) {
		assert.NoError(t, v.Validate(schema, map[string]interface{}{"active": true}))
	})

	t.Run("valid bool false", func(t *testing.T) {
		assert.NoError(t, v.Validate(schema, map[string]interface{}{"active": false}))
	})

	t.Run("string is not boolean", func(t *testing.T) {
		err := v.Validate(schema, map[string]interface{}{"active": "yes"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "active")
	})
}

func TestTemplateValidator_ValidateType_Object(t *testing.T) {
	v := NewTemplateValidator()
	schema := map[string]interface{}{
		"address": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"street": map[string]interface{}{"type": "string", "required": true},
				"zip":    map[string]interface{}{"type": "string"},
			},
		},
	}

	t.Run("valid nested object", func(t *testing.T) {
		data := map[string]interface{}{
			"address": map[string]interface{}{"street": "Main St", "zip": "12345"},
		}
		assert.NoError(t, v.Validate(schema, data))
	})

	t.Run("nested required field missing", func(t *testing.T) {
		data := map[string]interface{}{
			"address": map[string]interface{}{"zip": "12345"},
		}
		err := v.Validate(schema, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "street")
	})

	t.Run("object field but wrong type", func(t *testing.T) {
		data := map[string]interface{}{"address": "not-an-object"}
		err := v.Validate(schema, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "address")
	})

	t.Run("object without properties schema is valid", func(t *testing.T) {
		s := map[string]interface{}{
			"meta": map[string]interface{}{"type": "object"},
		}
		data := map[string]interface{}{"meta": map[string]interface{}{"key": "value"}}
		assert.NoError(t, v.Validate(s, data))
	})

	t.Run("nested property wrong type", func(t *testing.T) {
		data := map[string]interface{}{
			"address": map[string]interface{}{"street": 42, "zip": "12345"},
		}
		err := v.Validate(schema, data)
		require.Error(t, err)
	})
}

func TestTemplateValidator_ValidateType_Array(t *testing.T) {
	v := NewTemplateValidator()
	schema := map[string]interface{}{
		"tags": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{"type": "string"},
		},
	}

	t.Run("valid string array", func(t *testing.T) {
		data := map[string]interface{}{"tags": []interface{}{"a", "b", "c"}}
		assert.NoError(t, v.Validate(schema, data))
	})

	t.Run("not an array", func(t *testing.T) {
		data := map[string]interface{}{"tags": "not-an-array"}
		err := v.Validate(schema, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tags")
	})

	t.Run("array item wrong type", func(t *testing.T) {
		data := map[string]interface{}{"tags": []interface{}{"valid", 42}}
		err := v.Validate(schema, data)
		require.Error(t, err)
	})

	t.Run("array without items schema is valid", func(t *testing.T) {
		s := map[string]interface{}{
			"items": map[string]interface{}{"type": "array"},
		}
		data := map[string]interface{}{"items": []interface{}{1, 2, 3}}
		assert.NoError(t, v.Validate(s, data))
	})
}

func TestTemplateValidator_ValidateType_Unknown(t *testing.T) {
	v := NewTemplateValidator()
	schema := map[string]interface{}{
		"weird": map[string]interface{}{"type": "custom_unknown_type"},
	}
	err := v.Validate(schema, map[string]interface{}{"weird": "anything"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}

func TestTemplateValidator_ValidateWithSampleData(t *testing.T) {
	v := NewTemplateValidator()
	schema := map[string]interface{}{
		"name": map[string]interface{}{"type": "string", "required": true},
	}

	t.Run("valid sample data", func(t *testing.T) {
		assert.NoError(t, v.ValidateWithSampleData(schema, map[string]interface{}{"name": "Test"}))
	})

	t.Run("invalid sample data", func(t *testing.T) {
		err := v.ValidateWithSampleData(schema, map[string]interface{}{})
		require.Error(t, err)
	})
}

func TestTemplateValidator_ExtractRequiredFields(t *testing.T) {
	v := NewTemplateValidator()

	t.Run("flat schema required fields", func(t *testing.T) {
		schema := map[string]interface{}{
			"name":  map[string]interface{}{"type": "string", "required": true},
			"email": map[string]interface{}{"type": "string", "required": true},
			"age":   map[string]interface{}{"type": "number"},
		}
		fields := v.ExtractRequiredFields(schema)
		assert.Len(t, fields, 2)
		assert.Contains(t, fields, "name")
		assert.Contains(t, fields, "email")
	})

	t.Run("nested required fields", func(t *testing.T) {
		schema := map[string]interface{}{
			"address": map[string]interface{}{
				"type":     "object",
				"required": true,
				"properties": map[string]interface{}{
					"street": map[string]interface{}{"type": "string", "required": true},
					"city":   map[string]interface{}{"type": "string"},
				},
			},
		}
		fields := v.ExtractRequiredFields(schema)
		assert.Contains(t, fields, "address")
		assert.Contains(t, fields, "address.street")
	})

	t.Run("empty schema returns empty slice", func(t *testing.T) {
		fields := v.ExtractRequiredFields(map[string]interface{}{})
		assert.Empty(t, fields)
	})

	t.Run("non-map field schema is skipped", func(t *testing.T) {
		schema := map[string]interface{}{
			"bad": "not-a-map",
		}
		fields := v.ExtractRequiredFields(schema)
		assert.Empty(t, fields)
	})
}
