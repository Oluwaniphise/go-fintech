package validation

import (
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("password", validatePassword)
}

type FieldError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

func Validate(data any) []FieldError {
	err := validate.Struct(data)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []FieldError{
			{
				Field:   "request",
				Rule:    "invalid",
				Message: err.Error(),
			},
		}
	}

	errors := make([]FieldError, 0, len(validationErrors))

	for _, fieldErr := range validationErrors {
		field := lowerFirst(fieldErr.Field())

		errors = append(errors, FieldError{
			Field:   field,
			Rule:    fieldErr.Tag(),
			Message: messageFor(field, fieldErr),
		})
	}

	return errors
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}

	return strings.ToLower(value[:1]) + value[1:]
}

func messageFor(field string, err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "e164":
		return field + " must be a valid international phone number"
	case "min":
		return field + " is too short"
	case "max":
		return field + " is too long"
	case "password":
		return field + " must be at least 8 characters and include uppercase, lowercase, number, and special character"
	case "len":
		return field + " has invalid length"
	case "numeric":
		return field + " must contain only numbers"
	case "uuid4":
		return field + " must be a valid reference ID"
	default:
		return field + " is invalid"
	}
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}
