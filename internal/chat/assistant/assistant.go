package assistant

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/chat/tools"
	"github.com/openai/openai-go/v2"
)

type Assistant struct {
	cli      openai.Client
	registry *tools.Registry
}

func New() *Assistant {
	// Initialize tool registry
	registry := tools.NewRegistry()
	registry.Register(tools.NewWeatherTool())
	registry.Register(tools.NewDateTool())
	registry.Register(tools.NewHolidayTool())
	registry.Register(tools.NewCalculatorTool()) // Bonus tool

	return &Assistant{
		cli:      openai.NewClient(),
		registry: registry,
	}
}

func (a *Assistant) Title(ctx context.Context, conv *model.Conversation) (string, error) {
	if len(conv.Messages) == 0 {
		return "An empty conversation", nil
	}

	slog.InfoContext(ctx, "Generating title for conversation", "conversation_id", conv.ID)

	// Create system message with instruction, then add user messages
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("Generate a concise, descriptive title for the conversation based on the user's first message. The title should be a single line, no more than 80 characters, and should not include any special characters or emojis. Only return the title, nothing else."),
	}

	// Add only the first user message for title generation
	for _, m := range conv.Messages {
		if m.Role == model.RoleUser {
			msgs = append(msgs, openai.UserMessage(m.Content))
			break // Only need the first user message
		}
	}

	resp, err := a.cli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModelGPT4oMini, // Use faster, cheaper model for title generation
		Messages: msgs,
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		return "", errors.New("empty response from OpenAI for title generation")
	}

	title := resp.Choices[0].Message.Content
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.Trim(title, " \t\r\n-\"'")

	if len(title) > 80 {
		title = title[:80]
	}

	return title, nil
}

func (a *Assistant) Reply(ctx context.Context, conv *model.Conversation) (string, error) {
	if len(conv.Messages) == 0 {
		return "", errors.New("conversation has no messages")
	}

	slog.InfoContext(ctx, "Generating reply for conversation", "conversation_id", conv.ID)

	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a helpful, concise AI assistant. Provide accurate, safe, and clear responses."),
	}

	for _, m := range conv.Messages {
		switch m.Role {
		case model.RoleUser:
			msgs = append(msgs, openai.UserMessage(m.Content))
		case model.RoleAssistant:
			msgs = append(msgs, openai.AssistantMessage(m.Content))
		}
	}

	for i := 0; i < 15; i++ {
		resp, err := a.cli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModelGPT4_1,
			Messages: msgs,
			Tools:    a.registry.Definitions(), // Use registry for tool definitions
		})

		if err != nil {
			return "", err
		}

		if len(resp.Choices) == 0 {
			return "", errors.New("no choices returned by OpenAI")
		}

		if message := resp.Choices[0].Message; len(message.ToolCalls) > 0 {
			msgs = append(msgs, message.ToParam())

			for _, call := range message.ToolCalls {
				slog.InfoContext(ctx, "Tool call received", "name", call.Function.Name, "args", call.Function.Arguments)

				// Look up and execute tool
				tool, exists := a.registry.Get(call.Function.Name)
				if !exists {
					msgs = append(msgs, openai.ToolMessage("unknown tool: "+call.Function.Name, call.ID))
					continue
				}

				result, err := tool.Execute(ctx, call.Function.Arguments)
				if err != nil {
					msgs = append(msgs, openai.ToolMessage("tool execution failed: "+err.Error(), call.ID))
					continue
				}

				msgs = append(msgs, openai.ToolMessage(result, call.ID))
			}

			continue
		}

		return resp.Choices[0].Message.Content, nil
	}

	return "", errors.New("too many tool calls, unable to generate reply")
}
