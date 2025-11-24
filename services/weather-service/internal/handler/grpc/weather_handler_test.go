package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "toxictoast/services/weather-service/api/proto"
	"toxictoast/services/weather-service/internal/domain"
)

// MockWeatherUseCase is a mock implementation for testing
type MockWeatherUseCase struct {
	CurrentWeatherFunc func(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error)
	ForecastFunc       func(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error)
}

func (m *MockWeatherUseCase) GetCurrentWeather(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error) {
	if m.CurrentWeatherFunc != nil {
		return m.CurrentWeatherFunc(ctx, lat, lon, tz)
	}
	return nil, errors.New("not implemented")
}

func (m *MockWeatherUseCase) GetForecast(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error) {
	if m.ForecastFunc != nil {
		return m.ForecastFunc(ctx, lat, lon, days, tz)
	}
	return nil, errors.New("not implemented")
}

func TestNewWeatherHandler(t *testing.T) {
	mockUC := &MockWeatherUseCase{}
	handler := NewWeatherHandler(mockUC, "1.0.0")

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
	if handler.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", handler.version)
	}
}

func TestWeatherHandler_GetCurrentWeather(t *testing.T) {
	ctx := context.Background()

	t.Run("successful weather retrieval", func(t *testing.T) {
		now := time.Now()
		expectedWeather := &domain.CurrentWeather{
			Latitude:            50.1109,
			Longitude:           8.6821,
			Time:                now,
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

		mockUC := &MockWeatherUseCase{
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

		handler := NewWeatherHandler(mockUC, "1.0.0")
		req := &pb.WeatherRequest{
			Latitude:  50.1109,
			Longitude: 8.6821,
			Timezone:  "Europe/Berlin",
		}

		resp, err := handler.GetCurrentWeather(ctx, req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		if resp.Latitude != 50.1109 {
			t.Errorf("Expected latitude 50.1109, got %f", resp.Latitude)
		}
		if resp.Temperature != 15.5 {
			t.Errorf("Expected temperature 15.5, got %f", resp.Temperature)
		}
		if resp.WeatherCode != 2 {
			t.Errorf("Expected weather code 2, got %d", resp.WeatherCode)
		}
		if resp.WeatherDescription != "Partly cloudy" {
			t.Errorf("Expected description 'Partly cloudy', got %s", resp.WeatherDescription)
		}
	})

	t.Run("use case error", func(t *testing.T) {
		mockUC := &MockWeatherUseCase{
			CurrentWeatherFunc: func(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error) {
				return nil, errors.New("API error")
			},
		}

		handler := NewWeatherHandler(mockUC, "1.0.0")
		req := &pb.WeatherRequest{
			Latitude:  50.1109,
			Longitude: 8.6821,
			Timezone:  "Europe/Berlin",
		}

		resp, err := handler.GetCurrentWeather(ctx, req)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if resp != nil {
			t.Error("Expected nil response on error")
		}
		if err.Error() != "API error" {
			t.Errorf("Expected 'API error', got %s", err.Error())
		}
	})

	t.Run("different coordinates", func(t *testing.T) {
		mockUC := &MockWeatherUseCase{
			CurrentWeatherFunc: func(ctx context.Context, lat, lon float64, tz string) (*domain.CurrentWeather, error) {
				return &domain.CurrentWeather{
					Latitude:  lat,
					Longitude: lon,
				}, nil
			},
		}

		handler := NewWeatherHandler(mockUC, "1.0.0")
		req := &pb.WeatherRequest{
			Latitude:  52.5200,
			Longitude: 13.4050,
			Timezone:  "Europe/Berlin",
		}

		resp, err := handler.GetCurrentWeather(ctx, req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.Latitude != 52.5200 {
			t.Errorf("Expected latitude 52.5200, got %f", resp.Latitude)
		}
		if resp.Longitude != 13.4050 {
			t.Errorf("Expected longitude 13.4050, got %f", resp.Longitude)
		}
	})
}

func TestWeatherHandler_GetForecast(t *testing.T) {
	ctx := context.Background()

	t.Run("successful forecast retrieval", func(t *testing.T) {
		date := time.Date(2025, 11, 18, 0, 0, 0, 0, time.UTC)
		sunrise := time.Date(2025, 11, 18, 7, 30, 0, 0, time.UTC)
		sunset := time.Date(2025, 11, 18, 17, 0, 0, 0, time.UTC)

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
					Sunrise:                  sunrise,
					Sunset:                   sunset,
				},
			},
		}

		mockUC := &MockWeatherUseCase{
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

		handler := NewWeatherHandler(mockUC, "1.0.0")
		req := &pb.ForecastRequest{
			Latitude:  50.1109,
			Longitude: 8.6821,
			Days:      7,
			Timezone:  "Europe/Berlin",
		}

		resp, err := handler.GetForecast(ctx, req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		if resp.Latitude != 50.1109 {
			t.Errorf("Expected latitude 50.1109, got %f", resp.Latitude)
		}
		if len(resp.Daily) != 1 {
			t.Errorf("Expected 1 daily forecast, got %d", len(resp.Daily))
		}
		if resp.Daily[0].TemperatureMax != 18.5 {
			t.Errorf("Expected max temp 18.5, got %f", resp.Daily[0].TemperatureMax)
		}
		if resp.Daily[0].WeatherDescription != "Slight rain" {
			t.Errorf("Expected description 'Slight rain', got %s", resp.Daily[0].WeatherDescription)
		}
	})

	t.Run("use case error", func(t *testing.T) {
		mockUC := &MockWeatherUseCase{
			ForecastFunc: func(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error) {
				return nil, errors.New("Forecast API error")
			},
		}

		handler := NewWeatherHandler(mockUC, "1.0.0")
		req := &pb.ForecastRequest{
			Latitude:  50.1109,
			Longitude: 8.6821,
			Days:      7,
			Timezone:  "Europe/Berlin",
		}

		resp, err := handler.GetForecast(ctx, req)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if resp != nil {
			t.Error("Expected nil response on error")
		}
		if err.Error() != "Forecast API error" {
			t.Errorf("Expected 'Forecast API error', got %s", err.Error())
		}
	})

	t.Run("variable forecast days", func(t *testing.T) {
		mockUC := &MockWeatherUseCase{
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

		handler := NewWeatherHandler(mockUC, "1.0.0")

		// Test 3 days
		req3 := &pb.ForecastRequest{
			Latitude:  50.0,
			Longitude: 8.0,
			Days:      3,
			Timezone:  "Europe/Berlin",
		}
		resp3, err := handler.GetForecast(ctx, req3)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(resp3.Daily) != 3 {
			t.Errorf("Expected 3 daily forecasts, got %d", len(resp3.Daily))
		}

		// Test 7 days
		req7 := &pb.ForecastRequest{
			Latitude:  50.0,
			Longitude: 8.0,
			Days:      7,
			Timezone:  "Europe/Berlin",
		}
		resp7, err := handler.GetForecast(ctx, req7)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(resp7.Daily) != 7 {
			t.Errorf("Expected 7 daily forecasts, got %d", len(resp7.Daily))
		}
	})

	t.Run("empty forecast", func(t *testing.T) {
		mockUC := &MockWeatherUseCase{
			ForecastFunc: func(ctx context.Context, lat, lon float64, days int, tz string) (*domain.Forecast, error) {
				return &domain.Forecast{
					Latitude:  lat,
					Longitude: lon,
					Daily:     []domain.DailyForecast{},
				}, nil
			},
		}

		handler := NewWeatherHandler(mockUC, "1.0.0")
		req := &pb.ForecastRequest{
			Latitude:  50.0,
			Longitude: 8.0,
			Days:      0,
			Timezone:  "Europe/Berlin",
		}

		resp, err := handler.GetForecast(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(resp.Daily) != 0 {
			t.Errorf("Expected empty daily array, got %d", len(resp.Daily))
		}
	})
}

func TestWeatherHandler_HealthCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("health check returns healthy status", func(t *testing.T) {
		mockUC := &MockWeatherUseCase{}
		handler := NewWeatherHandler(mockUC, "1.0.0")
		req := &pb.HealthCheckRequest{}

		resp, err := handler.HealthCheck(ctx, req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		if resp.Status != "healthy" {
			t.Errorf("Expected status 'healthy', got %s", resp.Status)
		}
		if resp.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got %s", resp.Version)
		}
	})

	t.Run("health check with different version", func(t *testing.T) {
		mockUC := &MockWeatherUseCase{}
		handler := NewWeatherHandler(mockUC, "2.5.3")
		req := &pb.HealthCheckRequest{}

		resp, err := handler.HealthCheck(ctx, req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.Version != "2.5.3" {
			t.Errorf("Expected version '2.5.3', got %s", resp.Version)
		}
	})
}
