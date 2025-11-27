package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/weather-service/api/proto"
	"toxictoast/services/weather-service/internal/domain"
	"toxictoast/services/weather-service/internal/query"
)

type WeatherHandler struct {
	pb.UnimplementedWeatherServiceServer
	queryBus *cqrs.QueryBus
	version  string
}

func NewWeatherHandler(queryBus *cqrs.QueryBus, version string) *WeatherHandler {
	return &WeatherHandler{
		queryBus: queryBus,
		version:  version,
	}
}

func (h *WeatherHandler) GetCurrentWeather(ctx context.Context, req *pb.WeatherRequest) (*pb.CurrentWeatherResponse, error) {
	qry := &query.GetCurrentWeatherQuery{
		BaseQuery: cqrs.BaseQuery{},
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timezone:  req.Timezone,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	weather := result.(*domain.CurrentWeather)

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
	qry := &query.GetForecastQuery{
		BaseQuery: cqrs.BaseQuery{},
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Days:      int(req.Days),
		Timezone:  req.Timezone,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	forecast := result.(*domain.Forecast)

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
