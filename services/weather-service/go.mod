module toxictoast/services/weather-service

go 1.24.4

require (
	github.com/joho/godotenv v1.5.1
	github.com/toxictoast/toxictoastgo/shared v0.0.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
)

replace github.com/toxictoast/toxictoastgo/shared => ../../shared
