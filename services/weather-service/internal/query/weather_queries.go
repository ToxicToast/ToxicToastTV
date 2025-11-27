package query

import (
	"context"
	"errors"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/weather-service/pkg/openmeteo"
)

// ============================================================================
// Queries
// ============================================================================

type GetCurrentWeatherQuery struct {
	cqrs.BaseQuery
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

func (q *GetCurrentWeatherQuery) QueryName() string {
	return "get_current_weather"
}

func (q *GetCurrentWeatherQuery) Validate() error {
	if q.Latitude < -90 || q.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if q.Longitude < -180 || q.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	if q.Timezone == "" {
		return errors.New("timezone is required")
	}
	return nil
}

type GetForecastQuery struct {
	cqrs.BaseQuery
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Days      int     `json:"days"`
	Timezone  string  `json:"timezone"`
}

func (q *GetForecastQuery) QueryName() string {
	return "get_forecast"
}

func (q *GetForecastQuery) Validate() error {
	if q.Latitude < -90 || q.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if q.Longitude < -180 || q.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	if q.Days < 1 || q.Days > 16 {
		return errors.New("days must be between 1 and 16")
	}
	if q.Timezone == "" {
		return errors.New("timezone is required")
	}
	return nil
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetCurrentWeatherHandler struct {
	weatherClient *openmeteo.Client
}

func NewGetCurrentWeatherHandler(weatherClient *openmeteo.Client) *GetCurrentWeatherHandler {
	return &GetCurrentWeatherHandler{
		weatherClient: weatherClient,
	}
}

func (h *GetCurrentWeatherHandler) Handle(ctx context.Context, qry cqrs.Query) (interface{}, error) {
	query := qry.(*GetCurrentWeatherQuery)
	return h.weatherClient.GetCurrentWeather(ctx, query.Latitude, query.Longitude, query.Timezone)
}

type GetForecastHandler struct {
	weatherClient *openmeteo.Client
}

func NewGetForecastHandler(weatherClient *openmeteo.Client) *GetForecastHandler {
	return &GetForecastHandler{
		weatherClient: weatherClient,
	}
}

func (h *GetForecastHandler) Handle(ctx context.Context, qry cqrs.Query) (interface{}, error) {
	query := qry.(*GetForecastQuery)
	return h.weatherClient.GetForecast(ctx, query.Latitude, query.Longitude, query.Days, query.Timezone)
}
