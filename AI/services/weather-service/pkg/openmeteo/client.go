package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/httpclient"
	"toxictoast/services/weather-service/internal/domain"
)

const (
	baseURL = "https://api.open-meteo.com/v1"
)

// Client is the OpenMeteo API client
type Client struct {
	httpClient *httpclient.Client
}

// New creates a new OpenMeteo client
func New() *Client {
	config := httpclient.DefaultConfig()
	config.Timeout = 10 * time.Second
	config.MaxRetries = 3

	return &Client{
		httpClient: httpclient.New(config),
	}
}

// GetCurrentWeather fetches current weather from OpenMeteo
func (c *Client) GetCurrentWeather(ctx context.Context, latitude, longitude float64, timezone string) (*domain.CurrentWeather, error) {
	// Build URL with query parameters
	params := url.Values{}
	params.Add("latitude", fmt.Sprintf("%.6f", latitude))
	params.Add("longitude", fmt.Sprintf("%.6f", longitude))
	params.Add("current", "temperature_2m,apparent_temperature,weather_code,wind_speed_10m,wind_direction_10m,relative_humidity_2m,precipitation,cloud_cover,pressure_msl,visibility")

	if timezone != "" {
		params.Add("timezone", timezone)
	}

	apiURL := fmt.Sprintf("%s/forecast?%s", baseURL, params.Encode())

	// Make request
	data, err := c.httpClient.GetJSON(ctx, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather: %w", err)
	}

	// Parse response
	var response CurrentWeatherResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse weather response: %w", err)
	}

	// Convert to domain model
	weather := &domain.CurrentWeather{
		Latitude:            response.Latitude,
		Longitude:           response.Longitude,
		Temperature:         response.Current.Temperature,
		ApparentTemperature: response.Current.ApparentTemperature,
		WeatherCode:         response.Current.WeatherCode,
		WeatherDescription:  domain.GetWeatherDescription(response.Current.WeatherCode),
		WindSpeed:           response.Current.WindSpeed,
		WindDirection:       response.Current.WindDirection,
		Humidity:            response.Current.Humidity,
		Precipitation:       response.Current.Precipitation,
		CloudCover:          response.Current.CloudCover,
		Pressure:            response.Current.Pressure,
		Visibility:          response.Current.Visibility,
	}

	// Parse time
	if response.Current.Time != "" {
		t, err := time.Parse(time.RFC3339, response.Current.Time)
		if err == nil {
			weather.Time = t
		}
	}

	return weather, nil
}

// GetForecast fetches weather forecast from OpenMeteo
func (c *Client) GetForecast(ctx context.Context, latitude, longitude float64, days int, timezone string) (*domain.Forecast, error) {
	if days < 1 || days > 16 {
		days = 7 // Default to 7 days
	}

	// Build URL with query parameters
	params := url.Values{}
	params.Add("latitude", fmt.Sprintf("%.6f", latitude))
	params.Add("longitude", fmt.Sprintf("%.6f", longitude))
	params.Add("daily", "temperature_2m_max,temperature_2m_min,weather_code,precipitation_sum,precipitation_probability_max,wind_speed_10m_max,sunrise,sunset")
	params.Add("forecast_days", fmt.Sprintf("%d", days))

	if timezone != "" {
		params.Add("timezone", timezone)
	}

	apiURL := fmt.Sprintf("%s/forecast?%s", baseURL, params.Encode())

	// Make request
	data, err := c.httpClient.GetJSON(ctx, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch forecast: %w", err)
	}

	// Parse response
	var response ForecastResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse forecast response: %w", err)
	}

	// Convert to domain model
	forecast := &domain.Forecast{
		Latitude:  response.Latitude,
		Longitude: response.Longitude,
		Daily:     make([]domain.DailyForecast, 0, len(response.Daily.Time)),
	}

	for i := range response.Daily.Time {
		daily := domain.DailyForecast{
			TemperatureMax:           response.Daily.TemperatureMax[i],
			TemperatureMin:           response.Daily.TemperatureMin[i],
			WeatherCode:              response.Daily.WeatherCode[i],
			WeatherDescription:       domain.GetWeatherDescription(response.Daily.WeatherCode[i]),
			PrecipitationSum:         response.Daily.PrecipitationSum[i],
			PrecipitationProbability: response.Daily.PrecipitationProbability[i],
			WindSpeedMax:             response.Daily.WindSpeedMax[i],
		}

		// Parse date
		if date, err := time.Parse("2006-01-02", response.Daily.Time[i]); err == nil {
			daily.Date = date
		}

		// Parse sunrise
		if sunrise, err := time.Parse(time.RFC3339, response.Daily.Sunrise[i]); err == nil {
			daily.Sunrise = sunrise
		}

		// Parse sunset
		if sunset, err := time.Parse(time.RFC3339, response.Daily.Sunset[i]); err == nil {
			daily.Sunset = sunset
		}

		forecast.Daily = append(forecast.Daily, daily)
	}

	return forecast, nil
}
