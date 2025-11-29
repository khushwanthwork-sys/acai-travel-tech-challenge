package tools

import (
	"context"
	"time"

	"github.com/openai/openai-go/v2"
)

// DateTool provides current date and time
type DateTool struct{}

// NewDateTool creates a new date tool
func NewDateTool() *DateTool {
	return &DateTool{}
}

func (t *DateTool) Name() string {
	return "get_today_date"
}

func (t *DateTool) Definition() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        t.Name(),
		Description: openai.String("Get today's date and time in RFC3339 format"),
	})
}

func (t *DateTool) Execute(ctx context.Context, arguments string) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}
