package validation

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateRequired verifies required field validation
func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		expectErr bool
	}{
		{"valid string", "field", "value", false},
		{"empty string", "field", "", true},
		{"whitespace only", "field", "   ", true},
		{"valid with spaces", "field", " value ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.field, tt.value)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateRequired() error = %v, expectErr %v", err, tt.expectErr)
			}

			if err != nil && !errors.Is(err, ErrRequired) {
				t.Errorf("Expected ErrRequired, got %v", err)
			}
		})
	}
}

// TestValidateMaxLength verifies length validation
func TestValidateMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		maxLength int
		expectErr bool
	}{
		{"within limit", "field", "hello", 10, false},
		{"at limit", "field", "12345", 5, false},
		{"exceeds limit", "field", "toolong", 5, true},
		{"empty string", "field", "", 10, false},
		{"unicode chars", "field", "こんにちは", 10, false}, // 5 chars
		{"unicode exceeds", "field", "こんにちは世界", 5, true},  // 7 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxLength(tt.field, tt.value, tt.maxLength)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateMaxLength() error = %v, expectErr %v", err, tt.expectErr)
			}

			if err != nil && !errors.Is(err, ErrTooLong) {
				t.Errorf("Expected ErrTooLong, got %v", err)
			}
		})
	}
}

// TestValidateUTF8 verifies UTF-8 validation
func TestValidateUTF8(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		expectErr bool
	}{
		{"valid ascii", "field", "hello", false},
		{"valid unicode", "field", "こんにちは🌍", false},
		{"invalid utf8", "field", string([]byte{0xff, 0xfe, 0xfd}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUTF8(tt.field, tt.value)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateUTF8() error = %v, expectErr %v", err, tt.expectErr)
			}

			if err != nil && !errors.Is(err, ErrInvalidFormat) {
				t.Errorf("Expected ErrInvalidFormat, got %v", err)
			}
		})
	}
}

// TestValidateID verifies ID format validation
func TestValidateID(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		maxLength int
		expectErr bool
	}{
		{"valid alphanumeric", "id", "agent123", 20, false},
		{"valid with hyphens", "id", "my-agent-id", 20, false},
		{"valid with underscores", "id", "my_agent_id", 20, false},
		{"mixed valid chars", "id", "agent-123_test", 20, false},
		{"empty is valid", "id", "", 20, false}, // Empty handled by ValidateRequired
		{"invalid special chars", "id", "agent@123", 20, true},
		{"invalid spaces", "id", "agent 123", 20, true},
		{"invalid dots", "id", "agent.123", 20, true},
		{"exceeds max length", "id", "verylongagentidthatexceedslimit", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.field, tt.value, tt.maxLength)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateID() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

// TestValidateRunID verifies run ID validation
func TestValidateRunID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"valid run id", "run-1234567890", false},
		{"empty run id", "", true},
		{"invalid format", "run@123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRunID("run_id", tt.value)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateRunID() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

// TestValidateListLimit verifies list limit validation
func TestValidateListLimit(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		expectErr bool
	}{
		{"valid limit", 100, false},
		{"minimum limit", MinListLimit, false},
		{"maximum limit", MaxListLimit, false},
		{"below minimum", 0, true},
		{"above maximum", MaxListLimit + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListLimit(tt.limit)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateListLimit() error = %v, expectErr %v", err, tt.expectErr)
			}

			if err != nil && !errors.Is(err, ErrOutOfRange) {
				t.Errorf("Expected ErrOutOfRange, got %v", err)
			}
		})
	}
}

// TestValidateListOffset verifies list offset validation
func TestValidateListOffset(t *testing.T) {
	tests := []struct {
		name      string
		offset    int
		expectErr bool
	}{
		{"valid offset", 100, false},
		{"zero offset", 0, false},
		{"negative offset", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListOffset(tt.offset)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateListOffset() error = %v, expectErr %v", err, tt.expectErr)
			}

			if err != nil && !errors.Is(err, ErrOutOfRange) {
				t.Errorf("Expected ErrOutOfRange, got %v", err)
			}
		})
	}
}

// TestSanitizeString verifies string sanitization
func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "hello", "hello"},
		{"with leading space", "  hello", "hello"},
		{"with trailing space", "hello  ", "hello"},
		{"with both spaces", "  hello  ", "hello"},
		{"with null byte", "hello\x00world", "helloworld"},
		{"multiple null bytes", "a\x00b\x00c", "abc"},
		{"only spaces", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestCreateRunRequestValidate verifies CreateRunRequest validation
func TestCreateRunRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		req       CreateRunRequest
		expectErr bool
		errField  string
	}{
		{
			name: "valid request",
			req: CreateRunRequest{
				AgentID: "coding",
				Prompt:  "Write a hello world program",
			},
			expectErr: false,
		},
		{
			name: "missing agent_id",
			req: CreateRunRequest{
				Prompt: "Some prompt",
			},
			expectErr: true,
			errField:  "agent_id",
		},
		{
			name: "missing prompt",
			req: CreateRunRequest{
				AgentID: "coding",
			},
			expectErr: true,
			errField:  "prompt",
		},
		{
			name: "invalid agent_id format",
			req: CreateRunRequest{
				AgentID: "agent@invalid",
				Prompt:  "Some prompt",
			},
			expectErr: true,
			errField:  "agent_id",
		},
		{
			name: "prompt too long",
			req: CreateRunRequest{
				AgentID: "coding",
				Prompt:  strings.Repeat("a", MaxPromptLength+1),
			},
			expectErr: true,
			errField:  "prompt",
		},
		{
			name: "valid with resume_from_run_id",
			req: CreateRunRequest{
				AgentID:         "coding",
				Prompt:          "Continue previous task",
				ResumeFromRunID: "run-123",
			},
			expectErr: false,
		},
		{
			name: "invalid resume_from_run_id",
			req: CreateRunRequest{
				AgentID:         "coding",
				Prompt:          "Continue",
				ResumeFromRunID: "invalid@run",
			},
			expectErr: true,
			errField:  "resume_from_run_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("Validate() error = %v, expectErr %v", err, tt.expectErr)
			}

			if err != nil && tt.errField != "" {
				var valErr *ValidationError
				if errors.As(err, &valErr) {
					if valErr.Field != tt.errField {
						t.Errorf("Expected error field %q, got %q", tt.errField, valErr.Field)
					}
				}
			}
		})
	}
}

// TestAddInstructionRequestValidate verifies AddInstructionRequest validation
func TestAddInstructionRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		req       AddInstructionRequest
		expectErr bool
	}{
		{
			name:      "valid instruction",
			req:       AddInstructionRequest{Instruction: "Please continue"},
			expectErr: false,
		},
		{
			name:      "empty instruction",
			req:       AddInstructionRequest{Instruction: ""},
			expectErr: true,
		},
		{
			name:      "instruction too long",
			req:       AddInstructionRequest{Instruction: strings.Repeat("a", MaxInstructionLength+1)},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("Validate() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

// TestResumeRunRequestValidate verifies ResumeRunRequest validation
func TestResumeRunRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		req       ResumeRunRequest
		expectErr bool
	}{
		{
			name:      "valid message",
			req:       ResumeRunRequest{Message: "Continue with this message"},
			expectErr: false,
		},
		{
			name:      "empty message",
			req:       ResumeRunRequest{Message: ""},
			expectErr: true,
		},
		{
			name:      "message too long",
			req:       ResumeRunRequest{Message: strings.Repeat("a", MaxMessageLength+1)},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("Validate() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

// TestValidationError verifies ValidationError implementation
func TestValidationError(t *testing.T) {
	err := NewValidationError("test_field", "is invalid", ErrInvalidFormat)

	if err.Field != "test_field" {
		t.Errorf("Expected field 'test_field', got '%s'", err.Field)
	}

	if err.Message != "is invalid" {
		t.Errorf("Expected message 'is invalid', got '%s'", err.Message)
	}

	expectedStr := "test_field: is invalid"
	if err.Error() != expectedStr {
		t.Errorf("Expected error string '%s', got '%s'", expectedStr, err.Error())
	}

	if !errors.Is(err, ErrInvalidFormat) {
		t.Error("ValidationError should unwrap to underlying error")
	}
}
