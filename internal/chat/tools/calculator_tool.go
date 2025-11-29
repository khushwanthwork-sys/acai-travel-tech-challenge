package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/openai/openai-go/v2"
)

// CalculatorTool performs mathematical calculations
type CalculatorTool struct{}

// NewCalculatorTool creates a new calculator tool
func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
	return "calculate"
}

func (t *CalculatorTool) Definition() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        t.Name(),
		Description: openai.String("Evaluate mathematical expressions safely. Supports basic arithmetic, parentheses, and common math operations."),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]string{
					"type":        "string",
					"description": "Mathematical expression to evaluate (e.g., '2 + 2', '(10 * 5) / 2', 'sqrt(16)')",
				},
			},
			"required": []string{"expression"},
		},
	})
}

func (t *CalculatorTool) Execute(ctx context.Context, arguments string) (string, error) {
	var payload struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal([]byte(arguments), &payload); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	expression, err := govaluate.NewEvaluableExpression(payload.Expression)
	if err != nil {
		return "", fmt.Errorf("invalid expression: %w", err)
	}

	result, err := expression.Evaluate(nil)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate: %w", err)
	}

	return fmt.Sprintf("%v", result), nil
}
