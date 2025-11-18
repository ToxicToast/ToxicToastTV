package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"

	weatherpb "toxictoast/services/gateway-service/api/proto/weather"
)

// MirrorHandler handles mirror dashboard requests
type MirrorHandler struct {
	weatherClient weatherpb.WeatherServiceClient
}

// NewMirrorHandler creates a new mirror handler
func NewMirrorHandler(weatherConn *grpc.ClientConn) *MirrorHandler {
	return &MirrorHandler{
		weatherClient: weatherpb.NewWeatherServiceClient(weatherConn),
	}
}

// MirrorDashboardResponse is the aggregated response for the mirror
type MirrorDashboardResponse struct {
	Timestamp time.Time      `json:"timestamp"`
	Weather   *WeatherData   `json:"weather,omitempty"`
	Calendar  *CalendarData  `json:"calendar,omitempty"` // Future
	Shopping  *ShoppingData  `json:"shopping,omitempty"` // Future
}

// WeatherData contains weather information
type WeatherData struct {
	Current  *CurrentWeather  `json:"current"`
	Forecast []DailyForecast  `json:"forecast"`
}

// CurrentWeather contains current weather conditions
type CurrentWeather struct {
	Temperature        float64 `json:"temperature"`
	ApparentTemp       float64 `json:"apparentTemperature"`
	WeatherDescription string  `json:"weatherDescription"`
	WeatherCode        int     `json:"weatherCode"`
	WindSpeed          float64 `json:"windSpeed"`
	WindDirection      int     `json:"windDirection"`
	Humidity           int     `json:"humidity"`
	Precipitation      float64 `json:"precipitation"`
	CloudCover         int     `json:"cloudCover"`
}

// DailyForecast contains forecast for one day
type DailyForecast struct {
	Date                     string  `json:"date"`
	TempMax                  float64 `json:"temperatureMax"`
	TempMin                  float64 `json:"temperatureMin"`
	WeatherDescription       string  `json:"weatherDescription"`
	WeatherCode              int     `json:"weatherCode"`
	PrecipitationSum         float64 `json:"precipitationSum"`
	PrecipitationProbability float64 `json:"precipitationProbability"`
	WindSpeedMax             float64 `json:"windSpeedMax"`
	Sunrise                  string  `json:"sunrise"`
	Sunset                   string  `json:"sunset"`
}

// CalendarData - placeholder for future
type CalendarData struct {
	Events []interface{} `json:"events"`
}

// ShoppingData - placeholder for future
type ShoppingData struct {
	Items []interface{} `json:"items"`
}

// GetDashboard returns aggregated dashboard data
func (h *MirrorHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Get query parameters for location
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	timezone := r.URL.Query().Get("timezone")

	// Default to Frankfurt if not provided
	if latStr == "" {
		latStr = "50.1109"
	}
	if lonStr == "" {
		lonStr = "8.6821"
	}
	if timezone == "" {
		timezone = "Europe/Berlin"
	}

	// Parse coordinates
	var lat, lon float64
	if _, err := fmt.Sscanf(latStr, "%f", &lat); err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscanf(lonStr, "%f", &lon); err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response := &MirrorDashboardResponse{
		Timestamp: time.Now(),
	}

	// Fetch weather data
	weatherData, err := h.fetchWeather(ctx, lat, lon, timezone)
	if err != nil {
		log.Printf("Failed to fetch weather: %v", err)
		// Continue without weather data
	} else {
		response.Weather = weatherData
	}

	// TODO: Fetch calendar data when calendar-service is ready
	// TODO: Fetch shopping list from foodfolio-service

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// fetchWeather gets current weather and forecast
func (h *MirrorHandler) fetchWeather(ctx context.Context, lat, lon float64, timezone string) (*WeatherData, error) {
	weatherData := &WeatherData{}

	// Fetch current weather
	currentReq := &weatherpb.WeatherRequest{
		Latitude:  lat,
		Longitude: lon,
		Timezone:  timezone,
	}

	currentResp, err := h.weatherClient.GetCurrentWeather(ctx, currentReq)
	if err != nil {
		return nil, err
	}

	weatherData.Current = &CurrentWeather{
		Temperature:        currentResp.Temperature,
		ApparentTemp:       currentResp.ApparentTemperature,
		WeatherDescription: currentResp.WeatherDescription,
		WeatherCode:        int(currentResp.WeatherCode),
		WindSpeed:          currentResp.WindSpeed,
		WindDirection:      int(currentResp.WindDirection),
		Humidity:           int(currentResp.Humidity),
		Precipitation:      currentResp.Precipitation,
		CloudCover:         int(currentResp.CloudCover),
	}

	// Fetch forecast (3 days)
	forecastReq := &weatherpb.ForecastRequest{
		Latitude:  lat,
		Longitude: lon,
		Days:      3,
		Timezone:  timezone,
	}

	forecastResp, err := h.weatherClient.GetForecast(ctx, forecastReq)
	if err != nil {
		return nil, err
	}

	weatherData.Forecast = make([]DailyForecast, 0, len(forecastResp.Daily))
	for _, day := range forecastResp.Daily {
		forecast := DailyForecast{
			Date:                     day.Date.AsTime().Format("2006-01-02"),
			TempMax:                  day.TemperatureMax,
			TempMin:                  day.TemperatureMin,
			WeatherDescription:       day.WeatherDescription,
			WeatherCode:              int(day.WeatherCode),
			PrecipitationSum:         day.PrecipitationSum,
			PrecipitationProbability: day.PrecipitationProbability,
			WindSpeedMax:             day.WindSpeedMax,
			Sunrise:                  day.Sunrise.AsTime().Format(time.RFC3339),
			Sunset:                   day.Sunset.AsTime().Format(time.RFC3339),
		}
		weatherData.Forecast = append(weatherData.Forecast, forecast)
	}

	return weatherData, nil
}
