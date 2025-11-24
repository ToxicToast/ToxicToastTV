# Quick Start - User Service

## Voraussetzungen

- Go 1.21+
- PostgreSQL 15+
- Redpanda/Kafka (läuft auf localhost:19092)

## 1. Datenbank Setup

### Option A: Docker (Empfohlen)
```bash
docker run --name toxictoast-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=toxictoast \
  -p 5432:5432 \
  -d postgres:15
```

**Hinweis:** Diese Datenbank wird mit auth-service geteilt!

### Option B: Lokale PostgreSQL
```bash
createdb toxictoast
```

## 2. Redpanda/Kafka Setup

Wenn du Redpanda bereits in Docker laufen hast, ist dieser Schritt fertig!

Andernfalls:
```bash
docker run -d --name redpanda \
  -p 19092:19092 \
  -p 9644:9644 \
  docker.redpanda.com/redpandadata/redpanda:latest \
  redpanda start --smp 1 --memory 1G \
  --kafka-addr internal://0.0.0.0:9092,external://0.0.0.0:19092 \
  --advertise-kafka-addr internal://redpanda:9092,external://localhost:19092
```

## 3. Environment Configuration

Kopiere die `.env.example` und passe sie an:

```bash
cp .env.example .env
```

Minimale `.env` Konfiguration:
```env
# Service
SERVICE_NAME=user-service
PORT=8080
GRPC_PORT=9090

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=toxictoast

# Kafka
KAFKA_BROKERS=localhost:19092
```

## 4. Dependencies installieren

```bash
go mod download
# oder mit Makefile:
make deps
```

## 5. Service starten

```bash
go run cmd/server/main.go
# oder mit Makefile:
make run
```

Der Service startet auf:
- **gRPC**: `localhost:9090`

## 6. Service testen

### Mit grpcurl (Installation erforderlich)

```bash
# Liste alle Services auf
grpcurl -plaintext localhost:9090 list

# Erstelle einen User
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "username": "testuser",
  "password": "Test123!",
  "first_name": "Test",
  "last_name": "User"
}' localhost:9090 user.UserService/CreateUser

# Hole User by Email
grpcurl -plaintext -d '{
  "email": "test@example.com"
}' localhost:9090 user.UserService/GetUserByEmail

# Liste alle User
grpcurl -plaintext -d '{
  "page": 1,
  "page_size": 10
}' localhost:9090 user.UserService/ListUsers
```

### Mit Docker Compose

```bash
# Im Root-Verzeichnis des Projekts
docker-compose up user-service
```

## 7. Kafka Events testen

Der User-Service publiziert folgende Kafka Events:

- `user.created` - Neuer User erstellt
- `user.updated` - User aktualisiert
- `user.deleted` - User gelöscht
- `user.activated` - User aktiviert
- `user.deactivated` - User deaktiviert
- `user.password.changed` - Passwort geändert

### Events ansehen mit Redpanda Console

```bash
docker run -d --name redpanda-console \
  -p 8090:8080 \
  -e KAFKA_BROKERS=localhost:19092 \
  docker.redpanda.com/redpandadata/console:latest
```

Öffne dann http://localhost:8090 im Browser.

## Troubleshooting

### Datenbank Connection Fehler
- Stelle sicher, dass PostgreSQL läuft: `docker ps | grep postgres`
- Prüfe die Credentials in `.env`
- Teste die Connection: `psql -h localhost -U postgres -d toxictoast`

### Kafka Connection Fehler
- Stelle sicher, dass Redpanda läuft: `docker ps | grep redpanda`
- Der Service läuft weiter, auch wenn Kafka nicht verfügbar ist (Events werden nur geloggt)

### Port bereits in Verwendung
- Ändere `GRPC_PORT` in der `.env` Datei
- Oder stoppe den Service, der den Port bereits nutzt

## Entwicklung

### Build
```bash
make build
```

### Tests ausführen
```bash
make test
```

### Linting
```bash
make lint
```

### Proto-Files neu generieren
```bash
make proto-gen
```

## Produktions-Build

```bash
make build-prod
```

Dies erstellt ein statisch gelinkte Binary in `bin/user-service`.

## Integration mit anderen Services

### Auth Service
Der Auth-Service kommuniziert mit dem User-Service für:
- User-Registrierung
- Login (Email + Password Validierung)
- Token-Generierung mit User-Daten

### Notification Service
Kann `user.*` Events konsumieren für:
- Willkommens-E-Mails bei `user.created`
- Password-Reset bei `user.password.changed`

### SSE Service
Kann `user.*` Events an Browser-Clients streamen.

### Webhook Service
Kann `user.*` Events an externe Webhooks weiterleiten.
