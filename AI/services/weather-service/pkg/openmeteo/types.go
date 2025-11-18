package openmeteo

// CurrentWeatherResponse from OpenMeteo API
type CurrentWeatherResponse struct {
	Latitude  float64       `json:"latitude"`
	Longitude float64       `json:"longitude"`
	Current   CurrentValues `json:"current"`
}

// CurrentValues contains current weather values
type CurrentValues struct {
	Time                string  `json:"time"`
	Temperature         float64 `json:"temperature_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	WeatherCode         int     `json:"weather_code"`
	WindSpeed           float64 `json:"wind_speed_10m"`
	WindDirection       int     `json:"wind_direction_10m"`
	Humidity            int     `json:"relative_humidity_2m"`
	Precipitation       float64 `json:"precipitation"`
	CloudCover          int     `json:"cloud_cover"`
	Pressure            float64 `json:"pressure_msl"`
	Visibility          float64 `json:"visibility"`
}

// ForecastResponse from OpenMeteo API
type ForecastResponse struct {
	Latitude  float64     `json:"latitude"`
	Longitude float64     `json:"longitude"`
	Daily     DailyValues `json:"daily"`
}

// DailyValues contains daily forecast values
type DailyValues struct {
	Time                     []string  `json:"time"`
	TemperatureMax           []float64 `json:"temperature_2m_max"`
	TemperatureMin           []float64 `json:"temperature_2m_min"`
	WeatherCode              []int     `json:"weather_code"`
	PrecipitationSum         []float64 `json:"precipitation_sum"`
	PrecipitationProbability []float64 `json:"precipitation_probability_max"`
	WindSpeedMax             []float64 `json:"wind_speed_10m_max"`
	Sunrise                  []string  `json:"sunrise"`
	Sunset                   []string  `json:"sunset"`
}
