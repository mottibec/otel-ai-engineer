package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// CodingAgent is a specialized agent for coding tasks
type CodingAgent struct {
	*Agent
}

// NewCodingAgent creates a new coding agent with file system tools
func NewCodingAgent(client *anthropic.Client, logLevel config.LogLevel) (*CodingAgent, error) {
	systemPrompt := `You are an expert coding assistant with access to file system tools.

Your capabilities:
- Read, write, edit, and delete files
- Search for files using patterns
- List directory contents

When working on coding tasks:
1. First, understand the task by reading relevant files
2. Plan your changes carefully
3. Make incremental changes using the edit_file tool
4. Verify your changes by reading the file again if needed
5. Be precise and avoid making unnecessary changes

Always provide clear explanations of what you're doing and why.`

	agent := NewAgent(Config{
		Name:         "CodingAgent",
		Description:  "An AI agent specialized for coding tasks with file system access",
		Client:       client,
		Model:        anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:    4096,
		SystemPrompt: systemPrompt,
		LogLevel:     logLevel,
		Tools:        tools.GetFileSystemTools(),
	})

	return &CodingAgent{Agent: agent}, nil
}

// Code executes a coding task
func (ca *CodingAgent) Code(ctx context.Context, task string) *RunResult {
	return ca.Run(ctx, task)
}

// RefactorFile refactors a specific file
func (ca *CodingAgent) RefactorFile(ctx context.Context, filePath string, instructions string) *RunResult {
	prompt := fmt.Sprintf(`Refactor the file at path: %s

Instructions: %s

Please read the file, apply the refactoring, and save the changes.`, filePath, instructions)

	return ca.Run(ctx, prompt)
}

// ImplementFeature implements a new feature
func (ca *CodingAgent) ImplementFeature(ctx context.Context, description string, targetFiles []string) *RunResult {
	filesStr := ""
	if len(targetFiles) > 0 {
		filesStr = "\n\nTarget files to modify:\n"
		for _, f := range targetFiles {
			filesStr += fmt.Sprintf("- %s\n", f)
		}
	}

	prompt := fmt.Sprintf(`Implement the following feature:

%s%s

Please read the relevant files, implement the feature, and save your changes.`, description, filesStr)

	return ca.Run(ctx, prompt)
}

// FixBug attempts to fix a bug based on description
func (ca *CodingAgent) FixBug(ctx context.Context, bugDescription string, affectedFiles []string) *RunResult {
	filesStr := ""
	if len(affectedFiles) > 0 {
		filesStr = "\n\nAffected files:\n"
		for _, f := range affectedFiles {
			filesStr += fmt.Sprintf("- %s\n", f)
		}
	}

	prompt := fmt.Sprintf(`Fix the following bug:

%s%s

Please investigate the issue, identify the root cause, and implement a fix.`, bugDescription, filesStr)

	return ca.Run(ctx, prompt)
}
