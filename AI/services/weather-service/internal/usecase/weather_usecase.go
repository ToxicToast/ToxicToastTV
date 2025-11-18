package usecase

import (
	"context"
	"toxictoast/services/weather-service/internal/domain"
	"toxictoast/services/weather-service/pkg/openmeteo"
)

type WeatherUseCase interface {
	GetCurrentWeather(ctx context.Context, latitude, longitude float64, timezone string) (*domain.CurrentWeather, error)
	GetForecast(ctx context.Context, latitude, longitude float64, days int, timezone string) (*domain.Forecast, error)
}

type weatherUseCase struct {
	openMeteoClient *openmeteo.Client
}

func NewWeatherUseCase(openMeteoClient *openmeteo.Client) WeatherUseCase {
	return &weatherUseCase{
		openMeteoClient: openMeteoClient,
	}
}

func (uc *weatherUseCase) GetCurrentWeather(ctx context.Context, latitude, longitude float64, timezone string) (*domain.CurrentWeather, error) {
	return uc.openMeteoClient.GetCurrentWeather(ctx, latitude, longitude, timezone)
}

func (uc *weatherUseCase) GetForecast(ctx context.Context, latitude, longitude float64, days int, timezone string) (*domain.Forecast, error) {
	return uc.openMeteoClient.GetForecast(ctx, latitude, longitude, days, timezone)
}
