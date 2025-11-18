package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "toxictoast/services/weather-service/api/proto"
	"toxictoast/services/weather-service/internal/usecase"
)

type WeatherHandler struct {
	pb.UnimplementedWeatherServiceServer
	weatherUC usecase.WeatherUseCase
	version   string
}

func NewWeatherHandler(weatherUC usecase.WeatherUseCase, version string) *WeatherHandler {
	return &WeatherHandler{
		weatherUC: weatherUC,
		version:   version,
	}
}

func (h *WeatherHandler) GetCurrentWeather(ctx context.Context, req *pb.WeatherRequest) (*pb.CurrentWeatherResponse, error) {
	// Get weather from use case
	weather, err := h.weatherUC.GetCurrentWeather(ctx, req.Latitude, req.Longitude, req.Timezone)
	if err != nil {
		return nil, err
	}

	// Convert to proto response
	return &pb.CurrentWeatherResponse{
		Latitude:            weather.Latitude,
		Longitude:           weather.Longitude,
		Time:                timestamppb.New(weather.Time),
		Temperature:         weather.Temperature,
		ApparentTemperature: weather.ApparentTemperature,
		WeatherCode:         int32(weather.WeatherCode),
		WeatherDescription:  weather.WeatherDescription,
		WindSpeed:           weather.WindSpeed,
		WindDirection:       int32(weather.WindDirection),
		Humidity:            int32(weather.Humidity),
		Precipitation:       weather.Precipitation,
		CloudCover:          int32(weather.CloudCover),
		Pressure:            weather.Pressure,
		Visibility:          weather.Visibility,
	}, nil
}

func (h *WeatherHandler) GetForecast(ctx context.Context, req *pb.ForecastRequest) (*pb.ForecastResponse, error) {
	// Get forecast from use case
	forecast, err := h.weatherUC.GetForecast(ctx, req.Latitude, req.Longitude, int(req.Days), req.Timezone)
	if err != nil {
		return nil, err
	}

	// Convert to proto response
	daily := make([]*pb.DailyForecast, 0, len(forecast.Daily))
	for _, d := range forecast.Daily {
		daily = append(daily, &pb.DailyForecast{
			Date:                     timestamppb.New(d.Date),
			TemperatureMax:           d.TemperatureMax,
			TemperatureMin:           d.TemperatureMin,
			WeatherCode:              int32(d.WeatherCode),
			WeatherDescription:       d.WeatherDescription,
			PrecipitationSum:         d.PrecipitationSum,
			PrecipitationProbability: d.PrecipitationProbability,
			WindSpeedMax:             d.WindSpeedMax,
			Sunrise:                  timestamppb.New(d.Sunrise),
			Sunset:                   timestamppb.New(d.Sunset),
		})
	}

	return &pb.ForecastResponse{
		Latitude:  forecast.Latitude,
		Longitude: forecast.Longitude,
		Daily:     daily,
	}, nil
}

func (h *WeatherHandler) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:  "healthy",
		Version: h.version,
	}, nil
}
