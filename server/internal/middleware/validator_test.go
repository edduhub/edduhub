package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ValidationErrors ---

func TestValidationErrors_Error(t *testing.T) {
	ve := ValidationErrors{
		"name":  {"is required"},
		"email": {"must be valid", "is too long"},
	}

	msg := ve.Error()
	assert.Contains(t, msg, "name: is required")
	assert.Contains(t, msg, "email: must be valid, is too long")
}

// --- ValidateStruct: required ---

func TestValidateStruct_Required(t *testing.T) {
	type Input struct {
		Name string `json:"name" validate:"required"`
	}

	t.Run("passes when field is present", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: "Alice"})
		assert.NoError(t, err)
	})

	t.Run("fails when string field is empty", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: ""})
		require.Error(t, err)
		ve, ok := err.(ValidationErrors)
		require.True(t, ok)
		assert.Contains(t, ve["name"][0], "required")
	})
}

func TestValidateStruct_RequiredPointer(t *testing.T) {
	type Input struct {
		Name *string `json:"name" validate:"required"`
	}

	t.Run("fails when pointer is nil", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: nil})
		require.Error(t, err)
	})

	t.Run("passes when pointer is non-nil", func(t *testing.T) {
		s := "Alice"
		err := ValidateStruct(&Input{Name: &s})
		assert.NoError(t, err)
	})
}

func TestValidateStruct_RequiredSlice(t *testing.T) {
	type Input struct {
		Tags []string `json:"tags" validate:"required"`
	}

	t.Run("fails when slice is empty", func(t *testing.T) {
		err := ValidateStruct(&Input{Tags: []string{}})
		require.Error(t, err)
	})

	t.Run("passes when slice has elements", func(t *testing.T) {
		err := ValidateStruct(&Input{Tags: []string{"go"}})
		assert.NoError(t, err)
	})
}

// --- Min/Max ---

func TestValidateStruct_Min(t *testing.T) {
	type Input struct {
		Age int `json:"age" validate:"min=18"`
	}

	t.Run("passes at minimum", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Age: 18}))
	})

	t.Run("passes above minimum", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Age: 25}))
	})

	t.Run("fails below minimum", func(t *testing.T) {
		err := ValidateStruct(&Input{Age: 10})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["age"][0], "at least")
	})
}

func TestValidateStruct_Max(t *testing.T) {
	type Input struct {
		Score int `json:"score" validate:"max=100"`
	}

	t.Run("passes at maximum", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Score: 100}))
	})

	t.Run("fails above maximum", func(t *testing.T) {
		err := ValidateStruct(&Input{Score: 101})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["score"][0], "at most")
	})
}

func TestValidateStruct_MinFloat(t *testing.T) {
	type Input struct {
		Value float64 `json:"value" validate:"min=1.5"`
	}

	t.Run("passes when equal", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Value: 1.5}))
	})

	t.Run("fails below", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Value: 1.0}))
	})
}

func TestValidateStruct_MaxFloat(t *testing.T) {
	type Input struct {
		Value float64 `json:"value" validate:"max=99.9"`
	}

	t.Run("fails above", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Value: 100.0}))
	})
}

func TestValidateStruct_MinUint(t *testing.T) {
	type Input struct {
		Count uint `json:"count" validate:"min=5"`
	}

	t.Run("passes at minimum", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Count: 5}))
	})

	t.Run("fails below", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Count: 2}))
	})
}

func TestValidateStruct_MaxUint(t *testing.T) {
	type Input struct {
		Count uint `json:"count" validate:"max=10"`
	}

	t.Run("fails above", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Count: 20}))
	})
}

// --- MinLen/MaxLen/Len ---

func TestValidateStruct_MinLength(t *testing.T) {
	type Input struct {
		Name string `json:"name" validate:"minlen=3"`
	}

	t.Run("passes with sufficient length", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Name: "abc"}))
	})

	t.Run("fails when too short", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: "ab"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["name"][0], "at least 3 characters")
	})
}

func TestValidateStruct_MaxLength(t *testing.T) {
	type Input struct {
		Name string `json:"name" validate:"maxlen=5"`
	}

	t.Run("passes within limit", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Name: "hello"}))
	})

	t.Run("fails when too long", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: "hello!"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["name"][0], "at most 5 characters")
	})
}

func TestValidateStruct_MinLengthSlice(t *testing.T) {
	type Input struct {
		Items []string `json:"items" validate:"minlen=2"`
	}

	t.Run("passes with enough items", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Items: []string{"a", "b"}}))
	})

	t.Run("fails with too few items", func(t *testing.T) {
		err := ValidateStruct(&Input{Items: []string{"a"}})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["items"][0], "at least 2 items")
	})
}

func TestValidateStruct_MaxLengthSlice(t *testing.T) {
	type Input struct {
		Items []string `json:"items" validate:"maxlen=1"`
	}

	t.Run("fails with too many items", func(t *testing.T) {
		err := ValidateStruct(&Input{Items: []string{"a", "b"}})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["items"][0], "at most 1 items")
	})
}

func TestValidateStruct_ExactLength(t *testing.T) {
	type Input struct {
		Code string `json:"code" validate:"len=6"`
	}

	t.Run("passes with exact length", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Code: "abc123"}))
	})

	t.Run("fails with wrong length", func(t *testing.T) {
		err := ValidateStruct(&Input{Code: "abc"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["code"][0], "exactly 6 characters")
	})
}

func TestValidateStruct_ExactLengthSlice(t *testing.T) {
	type Input struct {
		Items []int `json:"items" validate:"len=3"`
	}

	t.Run("passes with exact item count", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Items: []int{1, 2, 3}}))
	})

	t.Run("fails with wrong item count", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Items: []int{1}}))
	})
}

// --- Email ---

func TestValidateStruct_Email(t *testing.T) {
	type Input struct {
		Email string `json:"email" validate:"email"`
	}

	t.Run("passes with valid email", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Email: "user@example.com"}))
	})

	t.Run("passes when empty (not required)", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Email: ""}))
	})

	t.Run("fails with invalid email", func(t *testing.T) {
		err := ValidateStruct(&Input{Email: "not-an-email"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["email"][0], "valid email")
	})

	t.Run("fails without domain", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Email: "user@"}))
	})
}

// --- URL ---

func TestValidateStruct_URL(t *testing.T) {
	type Input struct {
		Website string `json:"website" validate:"url"`
	}

	t.Run("passes with valid http URL", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Website: "http://example.com"}))
	})

	t.Run("passes with valid https URL", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Website: "https://example.com/path"}))
	})

	t.Run("passes when empty", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Website: ""}))
	})

	t.Run("fails with invalid URL", func(t *testing.T) {
		err := ValidateStruct(&Input{Website: "not-a-url"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["website"][0], "valid URL")
	})
}

// --- Numeric ---

func TestValidateStruct_Numeric(t *testing.T) {
	type Input struct {
		Phone string `json:"phone" validate:"numeric"`
	}

	t.Run("passes with digits", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Phone: "1234567890"}))
	})

	t.Run("passes when empty", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Phone: ""}))
	})

	t.Run("fails with non-numeric", func(t *testing.T) {
		err := ValidateStruct(&Input{Phone: "123abc"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["phone"][0], "numeric")
	})
}

// --- Alpha ---

func TestValidateStruct_Alpha(t *testing.T) {
	type Input struct {
		Name string `json:"name" validate:"alpha"`
	}

	t.Run("passes with letters", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Name: "Alice"}))
	})

	t.Run("passes when empty", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Name: ""}))
	})

	t.Run("fails with numbers", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: "Alice123"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["name"][0], "alphabetic")
	})
}

// --- Alphanumeric ---

func TestValidateStruct_Alphanumeric(t *testing.T) {
	type Input struct {
		Username string `json:"username" validate:"alphanumeric"`
	}

	t.Run("passes with letters and numbers", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Username: "user123"}))
	})

	t.Run("passes when empty", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Username: ""}))
	})

	t.Run("fails with special characters", func(t *testing.T) {
		err := ValidateStruct(&Input{Username: "user@123"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["username"][0], "alphanumeric")
	})
}

// --- Date ---

func TestValidateStruct_Date(t *testing.T) {
	type Input struct {
		Birthday string `json:"birthday" validate:"date"`
	}

	t.Run("passes with valid date", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Birthday: "2024-01-15"}))
	})

	t.Run("passes when empty", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Birthday: ""}))
	})

	t.Run("fails with invalid date", func(t *testing.T) {
		err := ValidateStruct(&Input{Birthday: "not-a-date"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["birthday"][0], "valid date")
	})
}

func TestValidateStruct_DateCustomFormat(t *testing.T) {
	type Input struct {
		Birthday string `json:"birthday" validate:"date=2006/01/02"`
	}

	t.Run("passes with custom format", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Birthday: "2024/01/15"}))
	})

	t.Run("fails with wrong format", func(t *testing.T) {
		err := ValidateStruct(&Input{Birthday: "2024-01-15"})
		require.Error(t, err)
	})
}

// --- gt/gte/lt/lte ---

func TestValidateStruct_GreaterThan(t *testing.T) {
	type Input struct {
		Value int `json:"value" validate:"gt=0"`
	}

	t.Run("passes when greater", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Value: 1}))
	})

	t.Run("fails when equal", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Value: 0}))
	})

	t.Run("fails when less", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Value: -1}))
	})
}

func TestValidateStruct_GreaterThanOrEqual(t *testing.T) {
	type Input struct {
		Value int `json:"value" validate:"gte=0"`
	}

	t.Run("passes when equal", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Value: 0}))
	})

	t.Run("passes when greater", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Value: 1}))
	})

	t.Run("fails when less", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Value: -1}))
	})
}

func TestValidateStruct_LessThan(t *testing.T) {
	type Input struct {
		Value int `json:"value" validate:"lt=100"`
	}

	t.Run("passes when less", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Value: 99}))
	})

	t.Run("fails when equal", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Value: 100}))
	})

	t.Run("fails when greater", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Value: 101}))
	})
}

func TestValidateStruct_LessThanOrEqual(t *testing.T) {
	type Input struct {
		Value int `json:"value" validate:"lte=100"`
	}

	t.Run("passes when equal", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Value: 100}))
	})

	t.Run("passes when less", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Value: 50}))
	})

	t.Run("fails when greater", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Value: 101}))
	})
}

func TestValidateStruct_CompareFloat(t *testing.T) {
	type Input struct {
		Price float64 `json:"price" validate:"gt=0,lte=9999.99"`
	}

	t.Run("passes within range", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Price: 19.99}))
	})

	t.Run("fails at zero", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Price: 0}))
	})

	t.Run("fails above max", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Price: 10000.0}))
	})
}

func TestValidateStruct_CompareUint(t *testing.T) {
	type Input struct {
		Count uint `json:"count" validate:"gt=0,lt=10"`
	}

	t.Run("passes within range", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Count: 5}))
	})

	t.Run("fails at zero", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Count: 0}))
	})
}

// --- OneOf ---

func TestValidateStruct_OneOf(t *testing.T) {
	type Input struct {
		Status string `json:"status" validate:"oneof=active inactive pending"`
	}

	t.Run("passes with valid value", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Status: "active"}))
	})

	t.Run("fails with invalid value", func(t *testing.T) {
		err := ValidateStruct(&Input{Status: "deleted"})
		require.Error(t, err)
		ve := err.(ValidationErrors)
		assert.Contains(t, ve["status"][0], "must be one of")
	})
}

func TestValidateStruct_OneOfInt(t *testing.T) {
	type Input struct {
		Priority int `json:"priority" validate:"oneof=1 2 3"`
	}

	t.Run("passes with valid int value", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Priority: 1}))
	})

	t.Run("fails with invalid int value", func(t *testing.T) {
		require.Error(t, ValidateStruct(&Input{Priority: 5}))
	})
}

// --- Omitempty ---

func TestValidateStruct_Omitempty(t *testing.T) {
	type Input struct {
		Name string `json:"name" validate:"omitempty,minlen=3"`
	}

	t.Run("validation still applies when value is present", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: "ab"})
		require.Error(t, err)
	})

	t.Run("passes with valid value", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Name: "abc"}))
	})
}

// --- Pointer fields (optional) ---

func TestValidateStruct_PointerFieldSkipped(t *testing.T) {
	type Input struct {
		Name *string `json:"name" validate:"minlen=3"`
	}

	t.Run("skips nil pointer without required", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Name: nil}))
	})

	t.Run("validates non-nil pointer value", func(t *testing.T) {
		s := "ab"
		err := ValidateStruct(&Input{Name: &s})
		require.Error(t, err)
	})

	t.Run("passes valid pointer value", func(t *testing.T) {
		s := "abc"
		assert.NoError(t, ValidateStruct(&Input{Name: &s}))
	})
}

// --- No validate tag ---

func TestValidateStruct_NoTag(t *testing.T) {
	type Input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	err := ValidateStruct(&Input{Name: "", Email: ""})
	assert.NoError(t, err)
}

// --- Non-struct input ---

func TestValidateStruct_NonStruct(t *testing.T) {
	s := "not a struct"
	err := ValidateStruct(&s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected struct")
}

// --- Multiple validation rules ---

func TestValidateStruct_MultipleRules(t *testing.T) {
	type Input struct {
		Name string `json:"name" validate:"required,minlen=3,maxlen=50"`
	}

	t.Run("fails required", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: ""})
		require.Error(t, err)
	})

	t.Run("fails minlen", func(t *testing.T) {
		err := ValidateStruct(&Input{Name: "ab"})
		require.Error(t, err)
	})

	t.Run("passes all rules", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&Input{Name: "Alice"}))
	})
}

// --- JSON field name fallback ---

func TestValidateStruct_FieldNameFallback(t *testing.T) {
	type Input struct {
		MyField string `validate:"required"`
	}

	err := ValidateStruct(&Input{MyField: ""})
	require.Error(t, err)
	ve := err.(ValidationErrors)
	// Uses struct field name when json tag is missing
	_, exists := ve["MyField"]
	assert.True(t, exists)
}
