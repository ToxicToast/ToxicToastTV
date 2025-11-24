module toxictoast/services/gateway-service

go 1.24.4

require (
	github.com/gorilla/mux v1.8.1
	github.com/prometheus/client_golang v1.23.2
	github.com/toxictoast/toxictoastgo/shared v0.0.0
	golang.org/x/time v0.14.0
	google.golang.org/grpc v1.77.0
	toxictoast/services/auth-service v0.0.0
	toxictoast/services/blog-service v0.0.0
	toxictoast/services/user-service v0.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/swaggo/files/v2 v2.0.0 // indirect
	github.com/swaggo/http-swagger/v2 v2.0.2
	github.com/swaggo/swag v1.8.1 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/protobuf v1.36.10
	gopkg.in/yaml.v2 v2.4.0 // indirect
	toxictoast/services/foodfolio-service v0.0.0
	toxictoast/services/link-service v0.0.0
	toxictoast/services/notification-service v0.0.0
	toxictoast/services/sse-service v0.0.0
	toxictoast/services/twitchbot-service v0.0.0
	toxictoast/services/warcraft-service v0.0.0
	toxictoast/services/webhook-service v0.0.0
)

replace github.com/toxictoast/toxictoastgo/shared => ../../shared

replace toxictoast/services/blog-service => ../blog-service

replace toxictoast/services/foodfolio-service => ../foodfolio-service

replace toxictoast/services/link-service => ../link-service

replace toxictoast/services/notification-service => ../notification-service

replace toxictoast/services/sse-service => ../sse-service

replace toxictoast/services/twitchbot-service => ../twitchbot-service

replace toxictoast/services/warcraft-service => ../warcraft-service

replace toxictoast/services/webhook-service => ../webhook-service

replace toxictoast/services/auth-service => ../auth-service

replace toxictoast/services/user-service => ../user-service
