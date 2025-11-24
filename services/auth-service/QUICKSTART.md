# Quick Start - Auth Service

## Voraussetzungen

- Go 1.21+
- PostgreSQL 15+
- Redpanda/Kafka (läuft auf localhost:19092)
- **User Service** muss laufen (für Registrierung & Login)

## 1. Datenbank Setup

### Option A: Docker (Empfohlen)
```bash
docker run --name toxictoast-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=toxictoast \
  -p 5432:5432 \
  -d postgres:15
```

**Hinweis:** Diese Datenbank wird mit user-service geteilt!

### Option B: Lokale PostgreSQL
```bash
createdb toxictoast
```

## 2. User Service starten

Auth-Service benötigt den User-Service. Starte ihn zuerst:

```bash
cd ../user-service
go run cmd/server/main.go
```

Oder mit Docker Compose:
```bash
docker-compose up user-service
```

## 3. Redpanda/Kafka Setup

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

## 4. Environment Configuration

Kopiere die `.env.example` und passe sie an:

```bash
cp .env.example .env
```

Minimale `.env` Konfiguration:
```env
# Service
SERVICE_NAME=auth-service
PORT=8080
GRPC_PORT=9090

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=toxictoast

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-me-in-production
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=168h

# Kafka
KAFKA_BROKERS=localhost:19092

# User Service
USER_SERVICE_ADDR=localhost:9090
```

**Wichtig:** Ändere `JWT_SECRET` in Production!

## 5. Dependencies installieren

```bash
go mod download
# oder mit Makefile:
make deps
```

## 6. Service starten

```bash
go run cmd/server/main.go
# oder mit Makefile:
make run
```

Der Service startet auf:
- **gRPC**: `localhost:9090`

## 7. Service testen

### Mit grpcurl (Installation erforderlich)

#### 1. Registrierung
```bash
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "username": "testuser",
  "password": "Test123!",
  "first_name": "Test",
  "last_name": "User"
}' localhost:9090 auth.AuthService/Register
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": "900"
}
```

#### 2. Login
```bash
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "password": "Test123!"
}' localhost:9090 auth.AuthService/Login
```

#### 3. Token Validierung
```bash
grpcurl -plaintext -d '{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}' localhost:9090 auth.AuthService/ValidateToken
```

**Response:**
```json
{
  "user_id": "uuid-here",
  "email": "test@example.com",
  "username": "testuser",
  "roles": ["user"],
  "permissions": ["read:own_profile"]
}
```

#### 4. Token Refresh
```bash
grpcurl -plaintext -d '{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}' localhost:9090 auth.AuthService/RefreshToken
```

### RBAC - Rollen & Permissions

#### Rolle erstellen
```bash
grpcurl -plaintext -d '{
  "name": "admin",
  "description": "Administrator role"
}' localhost:9090 auth.AuthService/CreateRole
```

#### Permission erstellen
```bash
grpcurl -plaintext -d '{
  "resource": "users",
  "action": "delete",
  "description": "Delete users"
}' localhost:9090 auth.AuthService/CreatePermission
```

#### Permission zu Rolle zuweisen
```bash
grpcurl -plaintext -d '{
  "role_id": "role-uuid",
  "permission_id": "permission-uuid"
}' localhost:9090 auth.AuthService/AssignPermissionToRole
```

#### Rolle zu User zuweisen
```bash
grpcurl -plaintext -d '{
  "user_id": "user-uuid",
  "role_id": "role-uuid"
}' localhost:9090 auth.AuthService/AssignRoleToUser
```

### Mit Docker Compose

```bash
# Im Root-Verzeichnis des Projekts
docker-compose up auth-service user-service
```

## 8. Kafka Events testen

Der Auth-Service publiziert folgende Kafka Events:

- `auth.registered` - User erfolgreich registriert
- `auth.login` - User erfolgreich eingeloggt
- `auth.token.refreshed` - Token wurde erneuert

### Events ansehen mit Redpanda Console

```bash
docker run -d --name redpanda-console \
  -p 8090:8080 \
  -e KAFKA_BROKERS=localhost:19092 \
  docker.redpanda.com/redpandadata/console:latest
```

Öffne dann http://localhost:8090 im Browser.

## Troubleshooting

### User Service nicht erreichbar
- Stelle sicher, dass user-service läuft: `grpcurl -plaintext localhost:9090 list`
- Prüfe `USER_SERVICE_ADDR` in `.env`

### Datenbank Connection Fehler
- Stelle sicher, dass PostgreSQL läuft: `docker ps | grep postgres`
- Prüfe die Credentials in `.env`
- Teste die Connection: `psql -h localhost -U postgres -d toxictoast`

### JWT Token Invalid
- Prüfe `JWT_SECRET` in der `.env`
- Stelle sicher, dass Access Token nicht abgelaufen ist (Standard: 15min)
- Verwende Refresh Token für neuen Access Token

### Kafka Connection Fehler
- Stelle sicher, dass Redpanda läuft: `docker ps | grep redpanda`
- Der Service läuft weiter, auch wenn Kafka nicht verfügbar ist (Events werden nur geloggt)

## JWT Token Details

### Access Token
- **Gültigkeit:** 15 Minuten (konfigurierbar via `JWT_ACCESS_DURATION`)
- **Enthält:** user_id, email, username, roles, permissions
- **Verwendung:** API-Authentifizierung

### Refresh Token
- **Gültigkeit:** 7 Tage (konfigurierbar via `JWT_REFRESH_DURATION`)
- **Enthält:** nur user_id
- **Verwendung:** Neue Access Tokens anfordern

### Token dekodieren (für Debugging)

```bash
# Online: https://jwt.io
# CLI:
echo "eyJhbGciOiJIUzI1NiIs..." | base64 -d
```

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

Dies erstellt ein statisch gelinkte Binary in `bin/auth-service`.

## Security Best Practices

### Für Produktion

1. **JWT_SECRET ändern:** Verwende einen starken, zufälligen Secret:
   ```bash
   openssl rand -base64 64
   ```

2. **HTTPS verwenden:** Niemals Tokens über HTTP übertragen

3. **Token Rotation:** Implementiere Token-Rotation für Refresh Tokens

4. **Rate Limiting:** Implementiere Rate Limiting für Login-Versuche

5. **Audit Logging:** Logge alle Auth-Events für Security Audits

## Integration mit anderen Services

### Gateway Service
Der Gateway kann Token-Validierung nutzen:
```go
// Middleware für JWT-Validierung
tokenResp, err := authClient.ValidateToken(ctx, &auth.ValidateTokenRequest{
    Token: accessToken,
})
```

### User Service
Auth-Service kommuniziert mit User-Service für:
- User-Registrierung via `CreateUser`
- Login via `GetUserByEmail` + `VerifyPassword`
- Token-Refresh via `GetUser`

### Frontend Integration
```javascript
// Login
const response = await fetch('/api/auth/login', {
  method: 'POST',
  body: JSON.stringify({ email, password })
});
const { access_token, refresh_token } = await response.json();

// API Call mit Token
fetch('/api/protected', {
  headers: {
    'Authorization': `Bearer ${access_token}`
  }
});
```

## RBAC Permissions Schema

### Standard Permissions

```
users:read:own     - User kann eigenes Profil lesen
users:write:own    - User kann eigenes Profil ändern
users:delete:own   - User kann eigenes Profil löschen
users:read:all     - Admin kann alle User lesen
users:write:all    - Admin kann alle User ändern
users:delete:all   - Admin kann alle User löschen
```

### Standard Roles

```
user    - Basis-Rolle (read:own, write:own)
admin   - Admin-Rolle (read:all, write:all, delete:all)
```
