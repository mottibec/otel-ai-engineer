package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// File system tool input structures

type ReadFileInput struct {
	Path string `json:"path"`
}

type WriteFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type EditFileInput struct {
	Path       string `json:"path"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
}

type DeleteFileInput struct {
	Path string `json:"path"`
}

type SearchFilesInput struct {
	Pattern   string `json:"pattern"`
	Directory string `json:"directory"`
}

type ListDirectoryInput struct {
	Path string `json:"path"`
}

// File system tool handlers

func ReadFile(input ReadFileInput) (interface{}, error) {
	if input.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	content, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]interface{}{
		"path":    input.Path,
		"content": string(content),
		"size":    len(content),
	}, nil
}

func WriteFile(input WriteFileInput) (interface{}, error) {
	if input.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(input.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	err := os.WriteFile(input.Path, []byte(input.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"path":          input.Path,
		"bytes_written": len(input.Content),
		"success":       true,
	}, nil
}

func EditFile(input EditFileInput) (interface{}, error) {
	if input.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
	if input.OldContent == "" {
		return nil, fmt.Errorf("old_content is required")
	}

	// Read current content
	content, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	currentContent := string(content)

	// Check if old content exists
	if !strings.Contains(currentContent, input.OldContent) {
		return nil, fmt.Errorf("old_content not found in file")
	}

	// Replace old content with new content
	newContent := strings.Replace(currentContent, input.OldContent, input.NewContent, 1)

	// Write updated content
	err = os.WriteFile(input.Path, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"path":    input.Path,
		"success": true,
		"changes": 1,
	}, nil
}

func DeleteFile(input DeleteFileInput) (interface{}, error) {
	if input.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// Check if file exists
	if _, err := os.Stat(input.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", input.Path)
	}

	// Delete file
	err := os.Remove(input.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return map[string]interface{}{
		"path":    input.Path,
		"success": true,
		"deleted": true,
	}, nil
}

func SearchFiles(input SearchFilesInput) (interface{}, error) {
	if input.Pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}
	if input.Directory == "" {
		input.Directory = "."
	}

	var matches []string

	err := filepath.Walk(input.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if filename matches pattern
		matched, err := filepath.Match(input.Pattern, filepath.Base(path))
		if err != nil {
			return err
		}

		if matched {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return map[string]interface{}{
		"pattern":   input.Pattern,
		"directory": input.Directory,
		"matches":   matches,
		"count":     len(matches),
	}, nil
}

func ListDirectory(input ListDirectoryInput) (interface{}, error) {
	if input.Path == "" {
		input.Path = "."
	}

	entries, err := os.ReadDir(input.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	files := make([]map[string]interface{}, 0)
	directories := make([]string, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		} else {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, map[string]interface{}{
				"name": entry.Name(),
				"size": info.Size(),
			})
		}
	}

	return map[string]interface{}{
		"path":        input.Path,
		"files":       files,
		"directories": directories,
		"file_count":  len(files),
		"dir_count":   len(directories),
	}, nil
}

// GetFileSystemTools returns an array of file system tool definitions
// These can be registered with a registry or passed directly to an agent
func GetFileSystemTools() []Tool {
	return []Tool{
		{
			Name:        "read_file",
			Description: "Read the contents of a file",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to read",
					},
				},
				Required: []string{"path"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input ReadFileInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return ReadFile(input)
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a file (creates file if it doesn't exist)",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to write",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to write to the file",
					},
				},
				Required: []string{"path", "content"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input WriteFileInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return WriteFile(input)
			},
		},
		{
			Name:        "edit_file",
			Description: "Edit a file by replacing old content with new content",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to edit",
					},
					"old_content": map[string]interface{}{
						"type":        "string",
						"description": "The exact content to be replaced",
					},
					"new_content": map[string]interface{}{
						"type":        "string",
						"description": "The new content to insert",
					},
				},
				Required: []string{"path", "old_content", "new_content"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input EditFileInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return EditFile(input)
			},
		},
		{
			Name:        "delete_file",
			Description: "Delete a file",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to delete",
					},
				},
				Required: []string{"path"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input DeleteFileInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return DeleteFile(input)
			},
		},
		{
			Name:        "search_files",
			Description: "Search for files matching a pattern in a directory",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "File name pattern to match (supports wildcards like *.go)",
					},
					"directory": map[string]interface{}{
						"type":        "string",
						"description": "Directory to search in (defaults to current directory)",
					},
				},
				Required: []string{"pattern"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input SearchFilesInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return SearchFiles(input)
			},
		},
		{
			Name:        "list_directory",
			Description: "List files and directories in a directory",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the directory to list (defaults to current directory)",
					},
				},
				Required: []string{},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input ListDirectoryInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return ListDirectory(input)
			},
		},
	}
}

// SetupFileSystemTools registers file system tools with the registry
// Deprecated: Use GetFileSystemTools() and register them directly instead
func SetupFileSystemTools(registry *ToolRegistry) error {
	tools := GetFileSystemTools()
	registry.RegisterTools(tools)
	return nil
}
