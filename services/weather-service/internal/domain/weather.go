package domain

import "time"

// CurrentWeather represents current weather conditions
type CurrentWeather struct {
	Latitude             float64
	Longitude            float64
	Time                 time.Time
	Temperature          float64
	ApparentTemperature  float64
	WeatherCode          int
	WeatherDescription   string
	WindSpeed            float64
	WindDirection        int
	Humidity             int
	Precipitation        float64
	CloudCover           int
	Pressure             float64
	Visibility           float64
}

// DailyForecast represents weather forecast for one day
type DailyForecast struct {
	Date                     time.Time
	TemperatureMax           float64
	TemperatureMin           float64
	WeatherCode              int
	WeatherDescription       string
	PrecipitationSum         float64
	PrecipitationProbability float64
	WindSpeedMax             float64
	Sunrise                  time.Time
	Sunset                   time.Time
}

// Forecast represents multi-day weather forecast
type Forecast struct {
	Latitude  float64
	Longitude float64
	Daily     []DailyForecast
}

// GetWeatherDescription returns a human-readable description for WMO weather code
func GetWeatherDescription(code int) string {
	descriptions := map[int]string{
		0:  "Clear sky",
		1:  "Mainly clear",
		2:  "Partly cloudy",
		3:  "Overcast",
		45: "Fog",
		48: "Depositing rime fog",
		51: "Light drizzle",
		53: "Moderate drizzle",
		55: "Dense drizzle",
		56: "Light freezing drizzle",
		57: "Dense freezing drizzle",
		61: "Slight rain",
		63: "Moderate rain",
		65: "Heavy rain",
		66: "Light freezing rain",
		67: "Heavy freezing rain",
		71: "Slight snow fall",
		73: "Moderate snow fall",
		75: "Heavy snow fall",
		77: "Snow grains",
		80: "Slight rain showers",
		81: "Moderate rain showers",
		82: "Violent rain showers",
		85: "Slight snow showers",
		86: "Heavy snow showers",
		95: "Thunderstorm",
		96: "Thunderstorm with slight hail",
		99: "Thunderstorm with heavy hail",
	}

	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return "Unknown"
}
