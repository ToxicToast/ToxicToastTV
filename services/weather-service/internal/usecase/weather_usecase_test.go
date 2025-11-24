package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"toxictoast/services/weather-service/internal/domain"
)

// MockWeatherClient is a mock implementation for testing
type MockWeatherClient struct {
	CurrentWeatherFunc func(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error)
	ForecastFunc       func(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error)
}

func (m *MockWeatherClient) GetCurrentWeather(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error) {
	if m.CurrentWeatherFunc != nil {
		return m.CurrentWeatherFunc(ctx, lat, lon, tz)
	}
	return nil, errors.New("not implemented")
}

func (m *MockWeatherClient) GetForecast(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error) {
	if m.ForecastFunc != nil {
		return m.ForecastFunc(ctx, lat, lon, days, tz)
	}
	return nil, errors.New("not implemented")
}

func TestWeatherUseCase_GetCurrentWeather(t *testing.T) {
	ctx := context.Background()

	t.Run("successful weather retrieval", func(t *testing.T) {
		expectedWeather := &domain.CurrentWeather{
			Latitude:            50.1109,
			Longitude:           8.6821,
			Time:                time.Now(),
			Temperature:         15.5,
			ApparentTemperature: 13.2,
			WeatherCode:         2,
			WeatherDescription:  "Partly cloudy",
			WindSpeed:           12.5,
			WindDirection:       180,
			Humidity:            75,
			Precipitation:       0.0,
			CloudCover:          50,
			Pressure:            1013.25,
			Visibility:          10000.0,
		}

		mockClient := &MockWeatherClient{
			CurrentWeatherFunc: func(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error) {
				if lat != 50.1109 {
					t.Errorf("Expected latitude 50.1109, got %f", lat)
				}
				if lon != 8.6821 {
					t.Errorf("Expected longitude 8.6821, got %f", lon)
				}
				if tz != "Europe/Berlin" {
					t.Errorf("Expected timezone Europe/Berlin, got %s", tz)
				}
				return expectedWeather, nil
			},
		}

		uc := NewWeatherUseCase(mockClient)
		weather, err := uc.GetCurrentWeather(ctx, 50.1109, 8.6821, "Europe/Berlin")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if weather == nil {
			t.Fatal("Expected weather data, got nil")
		}
		if weather.Temperature != 15.5 {
			t.Errorf("Expected temperature 15.5, got %f", weather.Temperature)
		}
		if weather.WeatherCode != 2 {
			t.Errorf("Expected weather code 2, got %d", weather.WeatherCode)
		}
	})

	t.Run("client error", func(t *testing.T) {
		mockClient := &MockWeatherClient{
			CurrentWeatherFunc: func(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error) {
				return nil, errors.New("API error")
			},
		}

		uc := NewWeatherUseCase(mockClient)
		_, err := uc.GetCurrentWeather(ctx, 50.1109, 8.6821, "Europe/Berlin")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != "API error" {
			t.Errorf("Expected 'API error', got %s", err.Error())
		}
	})

	t.Run("different coordinates", func(t *testing.T) {
		mockClient := &MockWeatherClient{
			CurrentWeatherFunc: func(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error) {
				return &domain.CurrentWeather{
					Latitude:  lat,
					Longitude: lon,
				}, nil
			},
		}

		uc := NewWeatherUseCase(mockClient)
		weather, err := uc.GetCurrentWeather(ctx, 52.5200, 13.4050, "Europe/Berlin")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if weather.Latitude != 52.5200 {
			t.Errorf("Expected latitude 52.5200, got %f", weather.Latitude)
		}
		if weather.Longitude != 13.4050 {
			t.Errorf("Expected longitude 13.4050, got %f", weather.Longitude)
		}
	})
}

func TestWeatherUseCase_GetForecast(t *testing.T) {
	ctx := context.Background()

	t.Run("successful forecast retrieval", func(t *testing.T) {
		date := time.Date(2025, 11, 18, 0, 0, 0, 0, time.UTC)
		expectedForecast := &domain.Forecast{
			Latitude:  50.1109,
			Longitude: 8.6821,
			Daily: []domain.DailyForecast{
				{
					Date:                     date,
					TemperatureMax:           18.5,
					TemperatureMin:           8.2,
					WeatherCode:              61,
					WeatherDescription:       "Slight rain",
					PrecipitationSum:         5.2,
					PrecipitationProbability: 70.0,
					WindSpeedMax:             25.5,
				},
			},
		}

		mockClient := &MockWeatherClient{
			ForecastFunc: func(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error) {
				if lat != 50.1109 {
					t.Errorf("Expected latitude 50.1109, got %f", lat)
				}
				if lon != 8.6821 {
					t.Errorf("Expected longitude 8.6821, got %f", lon)
				}
				if days != 7 {
					t.Errorf("Expected 7 days, got %d", days)
				}
				if tz != "Europe/Berlin" {
					t.Errorf("Expected timezone Europe/Berlin, got %s", tz)
				}
				return expectedForecast, nil
			},
		}

		uc := NewWeatherUseCase(mockClient)
		forecast, err := uc.GetForecast(ctx, 50.1109, 8.6821, 7, "Europe/Berlin")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if forecast == nil {
			t.Fatal("Expected forecast data, got nil")
		}
		if len(forecast.Daily) != 1 {
			t.Errorf("Expected 1 daily forecast, got %d", len(forecast.Daily))
		}
		if forecast.Daily[0].TemperatureMax != 18.5 {
			t.Errorf("Expected max temperature 18.5, got %f", forecast.Daily[0].TemperatureMax)
		}
	})

	t.Run("client error", func(t *testing.T) {
		mockClient := &MockWeatherClient{
			ForecastFunc: func(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error) {
				return nil, errors.New("Forecast API error")
			},
		}

		uc := NewWeatherUseCase(mockClient)
		_, err := uc.GetForecast(ctx, 50.1109, 8.6821, 7, "Europe/Berlin")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != "Forecast API error" {
			t.Errorf("Expected 'Forecast API error', got %s", err.Error())
		}
	})

	t.Run("variable forecast days", func(t *testing.T) {
		mockClient := &MockWeatherClient{
			ForecastFunc: func(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error) {
				daily := make([]domain.DailyForecast, days)
				for i := 0; i < days; i++ {
					daily[i] = domain.DailyForecast{
						Date:           time.Now().AddDate(0, 0, i),
						TemperatureMax: 20.0,
						TemperatureMin: 10.0,
					}
				}
				return &domain.Forecast{
					Latitude:  lat,
					Longitude: lon,
					Daily:     daily,
				}, nil
			},
		}

		uc := NewWeatherUseCase(mockClient)

		// Test 3 days
		forecast3, err := uc.GetForecast(ctx, 50.0, 8.0, 3, "Europe/Berlin")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(forecast3.Daily) != 3 {
			t.Errorf("Expected 3 daily forecasts, got %d", len(forecast3.Daily))
		}

		// Test 7 days
		forecast7, err := uc.GetForecast(ctx, 50.0, 8.0, 7, "Europe/Berlin")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(forecast7.Daily) != 7 {
			t.Errorf("Expected 7 daily forecasts, got %d", len(forecast7.Daily))
		}

		// Test 14 days
		forecast14, err := uc.GetForecast(ctx, 50.0, 8.0, 14, "Europe/Berlin")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(forecast14.Daily) != 14 {
			t.Errorf("Expected 14 daily forecasts, got %d", len(forecast14.Daily))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockClient := &MockWeatherClient{
			ForecastFunc: func(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error) {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				return &domain.Forecast{}, nil
			},
		}

		uc := NewWeatherUseCase(mockClient)

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := uc.GetForecast(ctx, 50.0, 8.0, 7, "Europe/Berlin")

		if err == nil {
			t.Fatal("Expected error due to cancelled context")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})
}

func TestNewWeatherUseCase(t *testing.T) {
	mockClient := &MockWeatherClient{}
	uc := NewWeatherUseCase(mockClient)

	if uc == nil {
		t.Fatal("Expected use case to be created, got nil")
	}

	// Verify it implements the interface
	var _ WeatherUseCase = uc
}
