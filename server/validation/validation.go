package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// Maximum lengths to prevent abuse
	MaxPromptLength      = 100000 // 100KB max prompt
	MaxInstructionLength = 10000  // 10KB max instruction
	MaxAgentIDLength     = 100
	MaxRunIDLength       = 100
	MaxMessageLength     = 50000 // 50KB max message

	// Query parameter limits
	MaxListLimit  = 1000
	MinListLimit  = 1
	MinListOffset = 0
)

var (
	// ErrRequired indicates a required field is missing
	ErrRequired = errors.New("field is required")

	// ErrTooLong indicates a field exceeds maximum length
	ErrTooLong = errors.New("field exceeds maximum length")

	// ErrInvalidFormat indicates invalid field format
	ErrInvalidFormat = errors.New("field has invalid format")

	// ErrOutOfRange indicates a value is out of acceptable range
	ErrOutOfRange = errors.New("value is out of acceptable range")

	// Regex for validating IDs (alphanumeric, hyphens, underscores)
	idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidationError wraps validation errors with field context
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// ValidateCreateRunRequest validates a create run request
type CreateRunRequest struct {
	AgentID         string
	Prompt          string
	ResumeFromRunID string
}

func (r *CreateRunRequest) Validate() error {
	// Validate AgentID
	if err := ValidateRequired("agent_id", r.AgentID); err != nil {
		return err
	}
	if err := ValidateID("agent_id", r.AgentID, MaxAgentIDLength); err != nil {
		return err
	}

	// Validate Prompt
	if err := ValidateRequired("prompt", r.Prompt); err != nil {
		return err
	}
	if err := ValidateMaxLength("prompt", r.Prompt, MaxPromptLength); err != nil {
		return err
	}
	if err := ValidateUTF8("prompt", r.Prompt); err != nil {
		return err
	}

	// Validate ResumeFromRunID if provided
	if r.ResumeFromRunID != "" {
		if err := ValidateID("resume_from_run_id", r.ResumeFromRunID, MaxRunIDLength); err != nil {
			return err
		}
	}

	return nil
}

// ValidateAddInstructionRequest validates an add instruction request
type AddInstructionRequest struct {
	Instruction string
}

func (r *AddInstructionRequest) Validate() error {
	if err := ValidateRequired("instruction", r.Instruction); err != nil {
		return err
	}
	if err := ValidateMaxLength("instruction", r.Instruction, MaxInstructionLength); err != nil {
		return err
	}
	if err := ValidateUTF8("instruction", r.Instruction); err != nil {
		return err
	}
	return nil
}

// ValidateResumeRunRequest validates a resume run request
type ResumeRunRequest struct {
	Message string
}

func (r *ResumeRunRequest) Validate() error {
	if err := ValidateRequired("message", r.Message); err != nil {
		return err
	}
	if err := ValidateMaxLength("message", r.Message, MaxMessageLength); err != nil {
		return err
	}
	if err := ValidateUTF8("message", r.Message); err != nil {
		return err
	}
	return nil
}

// ValidateRequired checks if a string field is non-empty
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(field, "is required", ErrRequired)
	}
	return nil
}

// ValidateMaxLength checks if a string doesn't exceed maximum length
func ValidateMaxLength(field, value string, maxLength int) error {
	length := utf8.RuneCountInString(value)
	if length > maxLength {
		return NewValidationError(
			field,
			fmt.Sprintf("exceeds maximum length of %d characters (got %d)", maxLength, length),
			ErrTooLong,
		)
	}
	return nil
}

// ValidateUTF8 checks if a string is valid UTF-8
func ValidateUTF8(field, value string) error {
	if !utf8.ValidString(value) {
		return NewValidationError(field, "contains invalid UTF-8 characters", ErrInvalidFormat)
	}
	return nil
}

// ValidateID validates an ID field (alphanumeric, hyphens, underscores)
func ValidateID(field, value string, maxLength int) error {
	if value == "" {
		return nil // Empty IDs are handled by ValidateRequired
	}

	if err := ValidateMaxLength(field, value, maxLength); err != nil {
		return err
	}

	if !idRegex.MatchString(value) {
		return NewValidationError(
			field,
			"must contain only alphanumeric characters, hyphens, and underscores",
			ErrInvalidFormat,
		)
	}

	return nil
}

// ValidateRunID validates a run ID
func ValidateRunID(field, value string) error {
	if err := ValidateRequired(field, value); err != nil {
		return err
	}
	return ValidateID(field, value, MaxRunIDLength)
}

// ValidateListLimit validates a list limit parameter
func ValidateListLimit(limit int) error {
	if limit < MinListLimit {
		return NewValidationError(
			"limit",
			fmt.Sprintf("must be at least %d", MinListLimit),
			ErrOutOfRange,
		)
	}
	if limit > MaxListLimit {
		return NewValidationError(
			"limit",
			fmt.Sprintf("cannot exceed %d", MaxListLimit),
			ErrOutOfRange,
		)
	}
	return nil
}

// ValidateListOffset validates a list offset parameter
func ValidateListOffset(offset int) error {
	if offset < MinListOffset {
		return NewValidationError(
			"offset",
			fmt.Sprintf("must be at least %d", MinListOffset),
			ErrOutOfRange,
		)
	}
	return nil
}

// SanitizeString removes potentially dangerous characters
// This is a basic sanitization - adjust based on your security requirements
func SanitizeString(value string) string {
	// Remove null bytes
	value = strings.ReplaceAll(value, "\x00", "")
	// Trim whitespace
	value = strings.TrimSpace(value)
	return value
}
