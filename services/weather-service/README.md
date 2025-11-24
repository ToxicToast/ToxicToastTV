# Weather Service

**Status:** ✅ Production Ready

A weather data service using the free Open-Meteo API to provide current weather conditions and forecasts.

## Features

- ✅ **Current Weather** - Real-time weather data (temperature, humidity, wind, precipitation)
- ✅ **Weather Forecast** - Up to 16 days forecast
- ✅ **Location-Based** - Queries using latitude/longitude coordinates
- ✅ **Timezone Support** - Automatic timezone handling
- ✅ **WMO Weather Codes** - Human-readable weather descriptions
- ✅ **Sunrise/Sunset** - Astronomical data included
- ✅ **No API Key Required** - Uses free Open-Meteo API
- ✅ **gRPC API** - High-performance Protocol Buffers interface
- ✅ **Health Checks** - Liveness and readiness endpoints

## Architecture

```
weather-service/
├── api/proto/          # gRPC Protocol Buffers definitions
├── cmd/server/         # Service entry point
├── internal/
│   ├── domain/         # Weather models
│   ├── usecase/        # Business logic
│   └── handler/        # gRPC service implementation
└── pkg/
    ├── config/         # Configuration management
    └── openmeteo/      # Open-Meteo API client
```

### Design Pattern

**Clean Architecture** with clear separation of concerns:
- **Domain Layer**: Pure weather data models
- **Use Case Layer**: Business logic for fetching and formatting weather data
- **Handler Layer**: gRPC service implementation
- **External Client**: Open-Meteo API integration (uses shared/httpclient)

## API

### gRPC Service

```protobuf
service WeatherService {
  rpc GetCurrentWeather(CurrentWeatherRequest) returns (CurrentWeatherResponse);
  rpc GetForecast(ForecastRequest) returns (ForecastResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

### Get Current Weather

**Request:**
```json
{
  "latitude": 50.1109,
  "longitude": 8.6821,
  "timezone": "Europe/Berlin"
}
```

**Response:**
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

### Get Forecast

**Request:**
```json
{
  "latitude": 50.1109,
  "longitude": 8.6821,
  "days": 7,
  "timezone": "Europe/Berlin"
}
```

**Response:**
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

The service uses **WMO Weather Interpretation Codes**:

| Code | Description |
|------|-------------|
| 0 | Clear sky |
| 1, 2, 3 | Mainly clear, partly cloudy, overcast |
| 45, 48 | Fog and depositing rime fog |
| 51, 53, 55 | Drizzle: Light, moderate, dense |
| 61, 63, 65 | Rain: Slight, moderate, heavy |
| 71, 73, 75 | Snow fall: Slight, moderate, heavy |
| 80, 81, 82 | Rain showers: Slight, moderate, violent |
| 85, 86 | Snow showers: Slight and heavy |
| 95, 96, 99 | Thunderstorm: Slight or moderate, with hail |

## Configuration

### Environment Variables

```env
# Service Configuration
SERVICE_NAME=weather-service
ENVIRONMENT=development

# Ports
PORT=8080
GRPC_PORT=9090

# Open-Meteo API (default: https://api.open-meteo.com)
OPENMETEO_BASE_URL=https://api.open-meteo.com
```

### Example Locations

| City | Latitude | Longitude |
|------|----------|-----------|
| Frankfurt | 50.1109 | 8.6821 |
| Berlin | 52.5200 | 13.4050 |
| Munich | 48.1351 | 11.5820 |
| Hamburg | 53.5511 | 9.9937 |
| New York | 40.7128 | -74.0060 |
| Tokyo | 35.6762 | 139.6503 |

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Internet connection (for Open-Meteo API)

### Running Locally

```bash
# Clone and navigate to service
cd services/weather-service

# Copy and configure environment
cp .env.example .env

# Run the service
go run cmd/server/main.go
```

The service will start:
- **gRPC Server**: `localhost:9090`
- **Health Check**: `http://localhost:8080/health`

### Using Make

```bash
# Build the service
make build

# Run the service
make run

# Run tests
make test

# Build Docker image
make docker-build

# Show all commands
make help
```

## Testing with grpcurl

### Get Current Weather

```bash
grpcurl -plaintext \
  -d '{"latitude": 50.1109, "longitude": 8.6821, "timezone": "Europe/Berlin"}' \
  localhost:9090 \
  weather.WeatherService/GetCurrentWeather
```

### Get 7-Day Forecast

```bash
grpcurl -plaintext \
  -d '{"latitude": 50.1109, "longitude": 8.6821, "days": 7, "timezone": "Europe/Berlin"}' \
  localhost:9090 \
  weather.WeatherService/GetForecast
```

### Health Check

```bash
grpcurl -plaintext \
  localhost:9090 \
  weather.WeatherService/HealthCheck
```

## Gateway Integration

The Weather Service is integrated into the Gateway Service:

- **Mirror Dashboard**: Available via `/api/mirror/dashboard`
- Aggregates data from Weather, Foodfolio, and Blog services
- Provides unified health status

## Data Source

This service uses [**Open-Meteo**](https://open-meteo.com/):
- ✅ Free weather API
- ✅ No API key required
- ✅ High-quality data from multiple sources
- ✅ Global coverage
- ✅ Real-time updates

**Attribution**: Weather data provided by Open-Meteo under CC BY 4.0 license.

## Development

### Project Structure

```
weather-service/
├── api/
│   └── proto/
│       ├── weather.proto          # gRPC service definition
│       ├── weather.pb.go          # Generated Go code
│       └── weather_grpc.pb.go     # Generated gRPC code
├── cmd/
│   └── server/
│       └── main.go                # Service entry point
├── internal/
│   ├── domain/
│   │   └── weather.go             # Weather models
│   ├── usecase/
│   │   └── weather_usecase.go     # Business logic
│   └── handler/
│       └── grpc/
│           └── weather_handler.go # gRPC implementation
├── pkg/
│   ├── config/
│   │   └── config.go              # Configuration
│   └── openmeteo/
│       └── client.go              # API client
├── Dockerfile                      # Container image
├── Makefile                        # Build automation
├── README.md                       # This file
└── QUICKSTART.md                   # Quick reference
```

### Building

```bash
# Build for current platform
go build -o bin/weather-service cmd/server/main.go

# Build for Linux (production)
GOOS=linux GOARCH=amd64 go build -o bin/weather-service-linux cmd/server/main.go
```

### Regenerating Proto Files

```bash
cd api/proto
protoc --go_out=. --go-grpc_out=. weather.proto
```

## Docker

### Build Image

```bash
docker build -t weather-service:latest .
```

### Run Container

```bash
docker run -p 8080:8080 -p 9090:9090 \
  -e GRPC_PORT=9090 \
  -e PORT=8080 \
  weather-service:latest
```

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "weather-service",
  "timestamp": "2025-11-17T20:00:00Z"
}
```

### Metrics

- Service uptime
- Request count
- Response times
- Error rates

## Error Handling

The service handles common errors:

| Error | Description | HTTP Status |
|-------|-------------|-------------|
| Invalid coordinates | Latitude/longitude out of range | 400 |
| API unavailable | Open-Meteo API timeout | 503 |
| Invalid timezone | Unknown timezone identifier | 400 |

## Performance

- **Average Response Time**: ~200-500ms (depends on Open-Meteo API)
- **Cache**: No caching (real-time data)
- **Rate Limiting**: Respects Open-Meteo API limits

## Future Enhancements

- [ ] Add caching layer for frequently requested locations
- [ ] Historical weather data
- [ ] Weather alerts and warnings
- [ ] Air quality index
- [ ] UV index
- [ ] Moon phase information
- [ ] Multi-day hourly forecasts

## Contributing

This is a private project. No external contributions accepted.

## License

Proprietary - ToxicToast

## Support

For issues or questions, check the QUICKSTART.md or service logs.

---

**Weather data provided by [Open-Meteo](https://open-meteo.com/)**
