package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Query Validation Tests
// ============================================================================

func TestGetCurrentWeatherQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   *GetCurrentWeatherQuery
		wantErr bool
	}{
		{
			name: "valid query",
			query: &GetCurrentWeatherQuery{
				Latitude:  52.52,
				Longitude: 13.405,
				Timezone:  "Europe/Berlin",
			},
			wantErr: false,
		},
		{
			name: "latitude too low",
			query: &GetCurrentWeatherQuery{
				Latitude:  -91,
				Longitude: 13.405,
				Timezone:  "Europe/Berlin",
			},
			wantErr: true,
		},
		{
			name: "latitude too high",
			query: &GetCurrentWeatherQuery{
				Latitude:  91,
				Longitude: 13.405,
				Timezone:  "Europe/Berlin",
			},
			wantErr: true,
		},
		{
			name: "longitude too low",
			query: &GetCurrentWeatherQuery{
				Latitude:  52.52,
				Longitude: -181,
				Timezone:  "Europe/Berlin",
			},
			wantErr: true,
		},
		{
			name: "longitude too high",
			query: &GetCurrentWeatherQuery{
				Latitude:  52.52,
				Longitude: 181,
				Timezone:  "Europe/Berlin",
			},
			wantErr: true,
		},
		{
			name: "missing timezone",
			query: &GetCurrentWeatherQuery{
				Latitude:  52.52,
				Longitude: 13.405,
				Timezone:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetCurrentWeatherQuery_QueryName(t *testing.T) {
	query := &GetCurrentWeatherQuery{}
	assert.Equal(t, "get_current_weather", query.QueryName())
}

func TestGetForecastQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   *GetForecastQuery
		wantErr bool
	}{
		{
			name: "valid query",
			query: &GetForecastQuery{
				Latitude:  52.52,
				Longitude: 13.405,
				Timezone:  "Europe/Berlin",
				Days:      7,
			},
			wantErr: false,
		},
		{
			name: "latitude too low",
			query: &GetForecastQuery{
				Latitude:  -91,
				Longitude: 13.405,
				Timezone:  "Europe/Berlin",
				Days:      7,
			},
			wantErr: true,
		},
		{
			name: "latitude too high",
			query: &GetForecastQuery{
				Latitude:  91,
				Longitude: 13.405,
				Timezone:  "Europe/Berlin",
				Days:      7,
			},
			wantErr: true,
		},
		{
			name: "longitude too low",
			query: &GetForecastQuery{
				Latitude:  52.52,
				Longitude: -181,
				Timezone:  "Europe/Berlin",
				Days:      7,
			},
			wantErr: true,
		},
		{
			name: "longitude too high",
			query: &GetForecastQuery{
				Latitude:  52.52,
				Longitude: 181,
				Timezone:  "Europe/Berlin",
				Days:      7,
			},
			wantErr: true,
		},
		{
			name: "missing timezone",
			query: &GetForecastQuery{
				Latitude:  52.52,
				Longitude: 13.405,
				Timezone:  "",
				Days:      7,
			},
			wantErr: true,
		},
		{
			name: "days too low",
			query: &GetForecastQuery{
				Latitude:  52.52,
				Longitude: 13.405,
				Timezone:  "Europe/Berlin",
				Days:      0,
			},
			wantErr: true,
		},
		{
			name: "days too high",
			query: &GetForecastQuery{
				Latitude:  52.52,
				Longitude: 13.405,
				Timezone:  "Europe/Berlin",
				Days:      17,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetForecastQuery_QueryName(t *testing.T) {
	query := &GetForecastQuery{}
	assert.Equal(t, "get_forecast", query.QueryName())
}

// Note: Handler tests omitted for weather-service as it uses external OpenMeteo API client
// which would require more complex mocking. Focus is on query validation which is the
// primary business logic for this read-only service.
//
// For integration tests, consider testing against a test OpenMeteo API instance
// or using a proper HTTP mock library like httptest.
