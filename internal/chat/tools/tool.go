package tools

import (
	"context"

	"github.com/openai/openai-go/v2"
)

// Tool represents a function that the assistant can call
type Tool interface {
	// Name returns the unique identifier for this tool
	Name() string

	// Definition returns the OpenAI function definition
	Definition() openai.ChatCompletionToolUnionParam

	// Execute runs the tool with given arguments and returns the result
	Execute(ctx context.Context, arguments string) (string, error)
}
