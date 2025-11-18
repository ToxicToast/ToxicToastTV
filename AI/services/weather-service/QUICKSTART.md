# Weather Service - Quick Start

Weather service using OpenMeteo API for weather data.

## Features

- ✅ Current weather conditions
- ✅ Weather forecast (up to 16 days)
- ✅ Location-based queries (latitude/longitude)
- ✅ Timezone support
- ✅ WMO weather code descriptions
- ✅ Sunrise/sunset times
- ✅ No API key required (uses free OpenMeteo API)

## API

### Get Current Weather

```bash
grpcurl -plaintext \
  -d '{"latitude": 50.1109, "longitude": 8.6821, "timezone": "Europe/Berlin"}' \
  localhost:9090 \
  weather.WeatherService/GetCurrentWeather
```

### Get Forecast

```bash
grpcurl -plaintext \
  -d '{"latitude": 50.1109, "longitude": 8.6821, "days": 7, "timezone": "Europe/Berlin"}' \
  localhost:9090 \
  weather.WeatherService/GetForecast
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Example Response - Current Weather

```json
{
  "latitude": 50.1109,
  "longitude": 8.6821,
  "time": "2025-11-17T20:00:00Z",
  "temperature": 12.5,
  "apparentTemperature": 10.2,
  "weatherCode": 2,
  "weatherDescription": "Partly cloudy",
  "windSpeed": 15.3,
  "windDirection": 240,
  "humidity": 75,
  "precipitation": 0.0,
  "cloudCover": 50,
  "pressure": 1013.5,
  "visibility": 10000
}
```

## Example Response - Forecast

```json
{
  "latitude": 50.1109,
  "longitude": 8.6821,
  "daily": [
    {
      "date": "2025-11-18T00:00:00Z",
      "temperatureMax": 15.2,
      "temperatureMin": 8.3,
      "weatherCode": 61,
      "weatherDescription": "Slight rain",
      "precipitationSum": 2.5,
      "precipitationProbability": 60,
      "windSpeedMax": 25.0,
      "sunrise": "2025-11-18T07:30:00Z",
      "sunset": "2025-11-18T17:00:00Z"
    }
  ]
}
```

## Weather Codes

The service uses WMO Weather interpretation codes:

- 0: Clear sky
- 1-3: Mainly clear to overcast
- 45-48: Fog
- 51-67: Various types of rain
- 71-77: Snow
- 80-82: Rain showers
- 85-86: Snow showers
- 95-99: Thunderstorm

## Locations

To get weather for a location, you need latitude and longitude:

- **Frankfurt**: 50.1109, 8.6821
- **Berlin**: 52.5200, 13.4050
- **Munich**: 48.1351, 11.5820
- **Hamburg**: 53.5511, 9.9937

## Environment Variables

```env
SERVICE_NAME=weather-service
ENVIRONMENT=development
PORT=8080
GRPC_PORT=9090
```

## Data Source

This service uses [Open-Meteo](https://open-meteo.com/), a free weather API that doesn't require an API key.

## Architecture

- **Domain**: Weather models
- **Use Case**: Business logic
- **OpenMeteo Client**: API integration (uses shared/httpclient)
- **gRPC Handler**: Service implementation
