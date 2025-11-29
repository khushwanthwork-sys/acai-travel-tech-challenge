package tools

import (
	"context"
	"encoding/json"

	"github.com/acai-travel/tech-challenge/internal/chat/weatherclient"
	"github.com/openai/openai-go/v2"
)

// WeatherTool provides weather information
type WeatherTool struct{}

// NewWeatherTool creates a new weather tool
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (t *WeatherTool) Name() string {
	return "get_weather"
}

func (t *WeatherTool) Definition() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        t.Name(),
		Description: openai.String("Get current weather and optional 3-day forecast for a given location"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]string{
					"type":        "string",
					"description": "City name, zip code, or coordinates (e.g., 'Barcelona', '10001', '48.8567,2.3508')",
				},
				"include_forecast": map[string]any{
					"type":        "boolean",
					"description": "Whether to include 3-day forecast. Default is false.",
					"default":     false,
				},
			},
			"required": []string{"location"},
		},
	})
}

func (t *WeatherTool) Execute(ctx context.Context, arguments string) (string, error) {
	var payload struct {
		Location        string `json:"location"`
		IncludeForecast bool   `json:"include_forecast,omitempty"`
	}

	if err := json.Unmarshal([]byte(arguments), &payload); err != nil {
		return "", err
	}

	// Use weatherclient package
	return weatherclient.GetWeather(ctx, payload.Location, payload.IncludeForecast)
}
