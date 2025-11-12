module toxictoast/services/gateway-service

go 1.24.4

require (
	github.com/gorilla/mux v1.8.1
	github.com/toxictoast/toxictoastgo/shared v0.0.0
	golang.org/x/time v0.14.0
	google.golang.org/grpc v1.76.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/toxictoast/toxictoastgo/shared => ../../shared
