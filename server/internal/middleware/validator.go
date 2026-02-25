package middleware

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// Validator interface for custom validation
type Validator interface {
	Validate() error
}

// ValidationErrors holds multiple validation errors
type ValidationErrors map[string][]string

func (v ValidationErrors) Error() string {
	var messages []string
	for field, errors := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", field, strings.Join(errors, ", ")))
	}
	return strings.Join(messages, "; ")
}

// ValidateRequest validates the request body and binds it to the target struct
func ValidateRequest(c echo.Context, target any) error {
	// Bind the request body
	if err := c.Bind(target); err != nil {
		return BadRequestError("Invalid request format", err)
	}

	// Perform validation
	if err := ValidateStruct(target); err != nil {
		if valErrs, ok := err.(ValidationErrors); ok {
			details := make(map[string]any)
			for field, errors := range valErrs {
				details[field] = errors
			}
			return ValidationError("Validation failed", details)
		}
		return BadRequestError("Validation failed", err)
	}

	// Check if target implements Validator interface for custom validation
	if validator, ok := target.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return BadRequestError("Custom validation failed", err)
		}
	}

	return nil
}

// ValidateStruct validates a struct based on tags
func ValidateStruct(s any) error {
	errors := make(ValidationErrors)

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", val.Kind())
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		tag := fieldType.Tag.Get("validate")

		if tag == "" {
			continue
		}

		jsonTag := fieldType.Tag.Get("json")
		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		// Skip validation if field is nil pointer and not required
		if field.Kind() == reflect.Pointer && field.IsNil() && !strings.Contains(tag, "required") {
			continue
		}

		fieldErrors := validateField(fieldName, field, tag)
		if len(fieldErrors) > 0 {
			errors[fieldName] = fieldErrors
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func validateField(name string, field reflect.Value, tag string) []string {
	var errors []string

	// Get actual value if it's a pointer
	actualValue := field
	if field.Kind() == reflect.Pointer && !field.IsNil() {
		actualValue = field.Elem()
	}

	rules := strings.SplitSeq(tag, ",")
	for rule := range rules {
		rule = strings.TrimSpace(rule)
		parts := strings.SplitN(rule, "=", 2)
		ruleName := parts[0]
		var ruleValue string
		if len(parts) > 1 {
			ruleValue = parts[1]
		}

		switch ruleName {
		case "required":
			if err := validateRequired(name, field); err != nil {
				errors = append(errors, err.Error())
			}
		case "min":
			if err := validateMin(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "max":
			if err := validateMax(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "minlen", "min_length":
			if err := validateMinLength(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "maxlen", "max_length":
			if err := validateMaxLength(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "len":
			if err := validateLength(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "email":
			if err := validateEmail(name, actualValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "url":
			if err := validateURL(name, actualValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "numeric":
			if err := validateNumeric(name, actualValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "alpha":
			if err := validateAlpha(name, actualValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "alphanumeric", "alphanum":
			if err := validateAlphanumeric(name, actualValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "date":
			if err := validateDate(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "gt":
			if err := validateGreaterThan(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "gte":
			if err := validateGreaterThanOrEqual(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "lt":
			if err := validateLessThan(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "lte":
			if err := validateLessThanOrEqual(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "oneof":
			if err := validateOneOf(name, actualValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		case "omitempty":
			// Skip validation if empty
			continue
		}
	}

	return errors
}

func validateRequired(name string, field reflect.Value) error {
	if field.Kind() == reflect.Pointer && field.IsNil() {
		return fmt.Errorf("%s is required", name)
	}

	if field.Kind() != reflect.Pointer {
		switch field.Kind() {
		case reflect.String:
			if field.String() == "" {
				return fmt.Errorf("%s is required", name)
			}
		case reflect.Slice, reflect.Map:
			if field.Len() == 0 {
				return fmt.Errorf("%s is required", name)
			}
		}
	}

	return nil
}

func validateMin(name string, field reflect.Value, min string) error {
	minVal, err := strconv.ParseFloat(min, 64)
	if err != nil {
		return fmt.Errorf("invalid min value: %s", min)
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(field.Int()) < minVal {
			return fmt.Errorf("%s must be at least %s", name, min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(field.Uint()) < minVal {
			return fmt.Errorf("%s must be at least %s", name, min)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() < minVal {
			return fmt.Errorf("%s must be at least %s", name, min)
		}
	}

	return nil
}

func validateMax(name string, field reflect.Value, max string) error {
	maxVal, err := strconv.ParseFloat(max, 64)
	if err != nil {
		return fmt.Errorf("invalid max value: %s", max)
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(field.Int()) > maxVal {
			return fmt.Errorf("%s must be at most %s", name, max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(field.Uint()) > maxVal {
			return fmt.Errorf("%s must be at most %s", name, max)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() > maxVal {
			return fmt.Errorf("%s must be at most %s", name, max)
		}
	}

	return nil
}

func validateMinLength(name string, field reflect.Value, minLen string) error {
	minLenVal, err := strconv.Atoi(minLen)
	if err != nil {
		return fmt.Errorf("invalid minlen value: %s", minLen)
	}

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) < minLenVal {
			return fmt.Errorf("%s must be at least %s characters long", name, minLen)
		}
	case reflect.Slice, reflect.Array:
		if field.Len() < minLenVal {
			return fmt.Errorf("%s must have at least %s items", name, minLen)
		}
	}

	return nil
}

func validateMaxLength(name string, field reflect.Value, maxLen string) error {
	maxLenVal, err := strconv.Atoi(maxLen)
	if err != nil {
		return fmt.Errorf("invalid maxlen value: %s", maxLen)
	}

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) > maxLenVal {
			return fmt.Errorf("%s must be at most %s characters long", name, maxLen)
		}
	case reflect.Slice, reflect.Array:
		if field.Len() > maxLenVal {
			return fmt.Errorf("%s must have at most %s items", name, maxLen)
		}
	}

	return nil
}

func validateLength(name string, field reflect.Value, length string) error {
	lenVal, err := strconv.Atoi(length)
	if err != nil {
		return fmt.Errorf("invalid len value: %s", length)
	}

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) != lenVal {
			return fmt.Errorf("%s must be exactly %s characters long", name, length)
		}
	case reflect.Slice, reflect.Array:
		if field.Len() != lenVal {
			return fmt.Errorf("%s must have exactly %s items", name, length)
		}
	}

	return nil
}

func validateEmail(name string, field reflect.Value) error {
	if field.Kind() != reflect.String {
		return nil
	}

	email := field.String()
	if email == "" {
		return nil // Use required tag to enforce non-empty
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%s must be a valid email address", name)
	}

	return nil
}

func validateURL(name string, field reflect.Value) error {
	if field.Kind() != reflect.String {
		return nil
	}

	url := field.String()
	if url == "" {
		return nil
	}

	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("%s must be a valid URL", name)
	}

	return nil
}

func validateNumeric(name string, field reflect.Value) error {
	if field.Kind() != reflect.String {
		return nil
	}

	str := field.String()
	if str == "" {
		return nil
	}

	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if !numericRegex.MatchString(str) {
		return fmt.Errorf("%s must contain only numeric characters", name)
	}

	return nil
}

func validateAlpha(name string, field reflect.Value) error {
	if field.Kind() != reflect.String {
		return nil
	}

	str := field.String()
	if str == "" {
		return nil
	}

	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(str) {
		return fmt.Errorf("%s must contain only alphabetic characters", name)
	}

	return nil
}

func validateAlphanumeric(name string, field reflect.Value) error {
	if field.Kind() != reflect.String {
		return nil
	}

	str := field.String()
	if str == "" {
		return nil
	}

	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumericRegex.MatchString(str) {
		return fmt.Errorf("%s must contain only alphanumeric characters", name)
	}

	return nil
}

func validateDate(name string, field reflect.Value, format string) error {
	if format == "" {
		format = "2006-01-02"
	}

	switch field.Kind() {
	case reflect.String:
		str := field.String()
		if str == "" {
			return nil
		}
		if _, err := time.Parse(format, str); err != nil {
			return fmt.Errorf("%s must be a valid date in format %s", name, format)
		}
	case reflect.Struct:
		if _, ok := field.Interface().(time.Time); !ok {
			return fmt.Errorf("%s must be a time.Time", name)
		}
	}

	return nil
}

func validateGreaterThan(name string, field reflect.Value, value string) error {
	compareVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid gt value: %s", value)
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(field.Int()) <= compareVal {
			return fmt.Errorf("%s must be greater than %s", name, value)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(field.Uint()) <= compareVal {
			return fmt.Errorf("%s must be greater than %s", name, value)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() <= compareVal {
			return fmt.Errorf("%s must be greater than %s", name, value)
		}
	}

	return nil
}

func validateGreaterThanOrEqual(name string, field reflect.Value, value string) error {
	compareVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid gte value: %s", value)
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(field.Int()) < compareVal {
			return fmt.Errorf("%s must be greater than or equal to %s", name, value)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(field.Uint()) < compareVal {
			return fmt.Errorf("%s must be greater than or equal to %s", name, value)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() < compareVal {
			return fmt.Errorf("%s must be greater than or equal to %s", name, value)
		}
	}

	return nil
}

func validateLessThan(name string, field reflect.Value, value string) error {
	compareVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid lt value: %s", value)
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(field.Int()) >= compareVal {
			return fmt.Errorf("%s must be less than %s", name, value)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(field.Uint()) >= compareVal {
			return fmt.Errorf("%s must be less than %s", name, value)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() >= compareVal {
			return fmt.Errorf("%s must be less than %s", name, value)
		}
	}

	return nil
}

func validateLessThanOrEqual(name string, field reflect.Value, value string) error {
	compareVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid lte value: %s", value)
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(field.Int()) > compareVal {
			return fmt.Errorf("%s must be less than or equal to %s", name, value)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(field.Uint()) > compareVal {
			return fmt.Errorf("%s must be less than or equal to %s", name, value)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() > compareVal {
			return fmt.Errorf("%s must be less than or equal to %s", name, value)
		}
	}

	return nil
}

func validateOneOf(name string, field reflect.Value, values string) error {
	allowedValues := strings.Split(values, " ")

	fieldValue := ""
	switch field.Kind() {
	case reflect.String:
		fieldValue = field.String()
	default:
		fieldValue = fmt.Sprintf("%v", field.Interface())
	}

	if slices.Contains(allowedValues, fieldValue) {
		return nil
	}

	return fmt.Errorf("%s must be one of: %s", name, strings.Join(allowedValues, ", "))
}

// ValidatorMiddleware creates a middleware that validates request bodies
func ValidatorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Store validator function in context for handlers to use
			c.Set("validator", ValidateStruct)
			return next(c)
		}
	}
}

// BindAndValidate is a helper function to bind and validate request
func BindAndValidate(c echo.Context, target any) error {
	if err := c.Bind(target); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format: "+err.Error())
	}

	if err := ValidateStruct(target); err != nil {
		if valErrs, ok := err.(ValidationErrors); ok {
			details := make(map[string]any)
			for field, errors := range valErrs {
				details[field] = errors
			}
			return ValidationError("Validation failed", details)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Check for custom validator
	if validator, ok := target.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Validation failed: "+err.Error())
		}
	}

	return nil
}
