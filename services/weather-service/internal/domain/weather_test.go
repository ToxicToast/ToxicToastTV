package domain

import (
	"testing"
	"time"
)

func TestGetWeatherDescription(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{"clear sky", 0, "Clear sky"},
		{"mainly clear", 1, "Mainly clear"},
		{"partly cloudy", 2, "Partly cloudy"},
		{"overcast", 3, "Overcast"},
		{"fog", 45, "Fog"},
		{"depositing rime fog", 48, "Depositing rime fog"},
		{"light drizzle", 51, "Light drizzle"},
		{"moderate drizzle", 53, "Moderate drizzle"},
		{"dense drizzle", 55, "Dense drizzle"},
		{"light freezing drizzle", 56, "Light freezing drizzle"},
		{"dense freezing drizzle", 57, "Dense freezing drizzle"},
		{"slight rain", 61, "Slight rain"},
		{"moderate rain", 63, "Moderate rain"},
		{"heavy rain", 65, "Heavy rain"},
		{"light freezing rain", 66, "Light freezing rain"},
		{"heavy freezing rain", 67, "Heavy freezing rain"},
		{"slight snow fall", 71, "Slight snow fall"},
		{"moderate snow fall", 73, "Moderate snow fall"},
		{"heavy snow fall", 75, "Heavy snow fall"},
		{"snow grains", 77, "Snow grains"},
		{"slight rain showers", 80, "Slight rain showers"},
		{"moderate rain showers", 81, "Moderate rain showers"},
		{"violent rain showers", 82, "Violent rain showers"},
		{"slight snow showers", 85, "Slight snow showers"},
		{"heavy snow showers", 86, "Heavy snow showers"},
		{"thunderstorm", 95, "Thunderstorm"},
		{"thunderstorm with slight hail", 96, "Thunderstorm with slight hail"},
		{"thunderstorm with heavy hail", 99, "Thunderstorm with heavy hail"},
		{"unknown code", 999, "Unknown"},
		{"negative code", -1, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWeatherDescription(tt.code)
			if result != tt.expected {
				t.Errorf("GetWeatherDescription(%d) = %s, expected %s", tt.code, result, tt.expected)
			}
		})
	}
}

func TestCurrentWeather_Structure(t *testing.T) {
	now := time.Now()
	weather := &CurrentWeather{
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

	if weather.Latitude != 50.1109 {
		t.Errorf("Expected latitude 50.1109, got %f", weather.Latitude)
	}
	if weather.Longitude != 8.6821 {
		t.Errorf("Expected longitude 8.6821, got %f", weather.Longitude)
	}
	if !weather.Time.Equal(now) {
		t.Errorf("Expected time %v, got %v", now, weather.Time)
	}
	if weather.Temperature != 15.5 {
		t.Errorf("Expected temperature 15.5, got %f", weather.Temperature)
	}
	if weather.WeatherCode != 2 {
		t.Errorf("Expected weather code 2, got %d", weather.WeatherCode)
	}
}

func TestDailyForecast_Structure(t *testing.T) {
	date := time.Date(2025, 11, 18, 0, 0, 0, 0, time.UTC)
	sunrise := time.Date(2025, 11, 18, 7, 30, 0, 0, time.UTC)
	sunset := time.Date(2025, 11, 18, 17, 0, 0, 0, time.UTC)

	forecast := &DailyForecast{
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
	}

	if !forecast.Date.Equal(date) {
		t.Errorf("Expected date %v, got %v", date, forecast.Date)
	}
	if forecast.TemperatureMax != 18.5 {
		t.Errorf("Expected max temp 18.5, got %f", forecast.TemperatureMax)
	}
	if forecast.TemperatureMin != 8.2 {
		t.Errorf("Expected min temp 8.2, got %f", forecast.TemperatureMin)
	}
	if forecast.WeatherCode != 61 {
		t.Errorf("Expected weather code 61, got %d", forecast.WeatherCode)
	}
	if forecast.PrecipitationProbability != 70.0 {
		t.Errorf("Expected precipitation probability 70.0, got %f", forecast.PrecipitationProbability)
	}
}

func TestForecast_Structure(t *testing.T) {
	date1 := time.Date(2025, 11, 18, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 11, 19, 0, 0, 0, 0, time.UTC)

	forecast := &Forecast{
		Latitude:  50.1109,
		Longitude: 8.6821,
		Daily: []DailyForecast{
			{
				Date:           date1,
				TemperatureMax: 15.0,
				TemperatureMin: 8.0,
				WeatherCode:    2,
			},
			{
				Date:           date2,
				TemperatureMax: 16.0,
				TemperatureMin: 9.0,
				WeatherCode:    0,
			},
		},
	}

	if forecast.Latitude != 50.1109 {
		t.Errorf("Expected latitude 50.1109, got %f", forecast.Latitude)
	}
	if forecast.Longitude != 8.6821 {
		t.Errorf("Expected longitude 8.6821, got %f", forecast.Longitude)
	}
	if len(forecast.Daily) != 2 {
		t.Errorf("Expected 2 daily forecasts, got %d", len(forecast.Daily))
	}
	if !forecast.Daily[0].Date.Equal(date1) {
		t.Errorf("Expected first date %v, got %v", date1, forecast.Daily[0].Date)
	}
	if !forecast.Daily[1].Date.Equal(date2) {
		t.Errorf("Expected second date %v, got %v", date2, forecast.Daily[1].Date)
	}
}

func TestWeatherCode_ValidCodes(t *testing.T) {
	validCodes := []int{0, 1, 2, 3, 45, 48, 51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 71, 73, 75, 77, 80, 81, 82, 85, 86, 95, 96, 99}

	for _, code := range validCodes {
		desc := GetWeatherDescription(code)
		if desc == "Unknown" {
			t.Errorf("Valid code %d returned 'Unknown'", code)
		}
	}
}

func TestWeatherCode_InvalidCodes(t *testing.T) {
	invalidCodes := []int{-1, 4, 10, 50, 100, 1000}

	for _, code := range invalidCodes {
		desc := GetWeatherDescription(code)
		if desc != "Unknown" {
			t.Errorf("Invalid code %d should return 'Unknown', got %s", code, desc)
		}
	}
}

func TestCurrentWeather_ZeroValues(t *testing.T) {
	weather := &CurrentWeather{}

	if weather.Latitude != 0.0 {
		t.Error("Zero value latitude should be 0.0")
	}
	if weather.Temperature != 0.0 {
		t.Error("Zero value temperature should be 0.0")
	}
	if weather.WeatherCode != 0 {
		t.Error("Zero value weather code should be 0")
	}
	if weather.Humidity != 0 {
		t.Error("Zero value humidity should be 0")
	}
}

func TestForecast_EmptyDaily(t *testing.T) {
	forecast := &Forecast{
		Latitude:  50.0,
		Longitude: 8.0,
		Daily:     []DailyForecast{},
	}

	if len(forecast.Daily) != 0 {
		t.Errorf("Expected empty daily array, got length %d", len(forecast.Daily))
	}
}

func TestDailyForecast_TemperatureRange(t *testing.T) {
	forecast := &DailyForecast{
		TemperatureMax: 25.0,
		TemperatureMin: 15.0,
	}

	if forecast.TemperatureMax <= forecast.TemperatureMin {
		t.Error("TemperatureMax should be greater than TemperatureMin")
	}

	tempRange := forecast.TemperatureMax - forecast.TemperatureMin
	if tempRange != 10.0 {
		t.Errorf("Expected temperature range 10.0, got %f", tempRange)
	}
}
