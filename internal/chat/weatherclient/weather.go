package weatherclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

// Response represents the response from WeatherAPI
type Response struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		TempF     float64 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
		WindKph   float64 `json:"wind_kph"`
		WindMph   float64 `json:"wind_mph"`
		Humidity  int     `json:"humidity"`
		FeelsLike float64 `json:"feelslike_c"`
	} `json:"current"`
	Forecast *struct {
		Forecastday []struct {
			Date string `json:"date"`
			Day  struct {
				MaxTempC  float64 `json:"maxtemp_c"`
				MinTempC  float64 `json:"mintemp_c"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
				ChanceOfRain int `json:"daily_chance_of_rain"`
			} `json:"day"`
		} `json:"forecastday"`
	} `json:"forecast,omitempty"`
}

// GetWeather fetches current weather for a location
func GetWeather(ctx context.Context, location string, includeForecast bool) (string, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("WEATHER_API_KEY environment variable not set")
	}

	// Build URL
	baseURL := "https://api.weatherapi.com/v1/current.json"
	if includeForecast {
		baseURL = "https://api.weatherapi.com/v1/forecast.json"
	}

	params := url.Values{}
	params.Add("key", apiKey)
	params.Add("q", location)
	params.Add("aqi", "no")
	if includeForecast {
		params.Add("days", "3")
	}

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	slog.InfoContext(ctx, "Fetching weather data", "location", location, "forecast", includeForecast)

	// Make request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("weather API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var weatherResp Response
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return "", fmt.Errorf("failed to parse weather response: %w", err)
	}

	// Format response
	result := fmt.Sprintf("Weather in %s, %s:\n", weatherResp.Location.Name, weatherResp.Location.Country)
	result += fmt.Sprintf("Temperature: %.1f°C (%.1f°F), feels like %.1f°C\n",
		weatherResp.Current.TempC, weatherResp.Current.TempF, weatherResp.Current.FeelsLike)
	result += fmt.Sprintf("Conditions: %s\n", weatherResp.Current.Condition.Text)
	result += fmt.Sprintf("Wind: %.1f km/h (%.1f mph)\n", weatherResp.Current.WindKph, weatherResp.Current.WindMph)
	result += fmt.Sprintf("Humidity: %d%%", weatherResp.Current.Humidity)

	// Add forecast if requested
	if includeForecast && weatherResp.Forecast != nil {
		result += "\n\n3-Day Forecast:\n"
		for _, day := range weatherResp.Forecast.Forecastday {
			result += fmt.Sprintf("%s: %s, High: %.1f°C, Low: %.1f°C, Rain chance: %d%%\n",
				day.Date, day.Day.Condition.Text, day.Day.MaxTempC, day.Day.MinTempC, day.Day.ChanceOfRain)
		}
	}

	return result, nil
}
