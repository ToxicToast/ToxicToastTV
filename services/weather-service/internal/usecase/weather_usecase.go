package usecase

import (
	"context"
	"toxictoast/services/weather-service/internal/domain"
	"toxictoast/services/weather-service/pkg/openmeteo"
)

// WeatherClient defines the interface for weather data clients
type WeatherClient interface {
	GetCurrentWeather(ctx context.Context, latitude, longitude float64, timezone string) (*domain.CurrentWeather, error)
	GetForecast(ctx context.Context, latitude, longitude float64, days int, timezone string) (*domain.Forecast, error)
}

type WeatherUseCase interface {
	GetCurrentWeather(ctx context.Context, latitude, longitude float64, timezone string) (*domain.CurrentWeather, error)
	GetForecast(ctx context.Context, latitude, longitude float64, days int, timezone string) (*domain.Forecast, error)
}

type weatherUseCase struct {
	weatherClient WeatherClient
}

func NewWeatherUseCase(weatherClient WeatherClient) WeatherUseCase {
	return &weatherUseCase{
		weatherClient: weatherClient,
	}
}

// NewWeatherUseCaseWithOpenMeteo creates a weather use case with OpenMeteo client
func NewWeatherUseCaseWithOpenMeteo(openMeteoClient *openmeteo.Client) WeatherUseCase {
	return &weatherUseCase{
		weatherClient: openMeteoClient,
	}
}

func (uc *weatherUseCase) GetCurrentWeather(ctx context.Context, latitude, longitude float64, timezone string) (*domain.CurrentWeather, error) {
	return uc.weatherClient.GetCurrentWeather(ctx, latitude, longitude, timezone)
}

func (uc *weatherUseCase) GetForecast(ctx context.Context, latitude, longitude float64, days int, timezone string) (*domain.Forecast, error) {
	return uc.weatherClient.GetForecast(ctx, latitude, longitude, days, timezone)
}
