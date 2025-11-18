# Keycloak Setup Guide für ToxicToastGo

Dieses Tutorial zeigt dir Schritt für Schritt, wie du Keycloak für dein Microservices-System einrichtest.

## Inhaltsverzeichnis

1. [Docker Compose Setup](#1-docker-compose-setup)
2. [Keycloak Admin Console](#2-keycloak-admin-console)
3. [Realm erstellen](#3-realm-erstellen)
4. [Clients konfigurieren](#4-clients-konfigurieren)
5. [Rollen (RBAC) einrichten](#5-rollen-rbac-einrichten)
6. [Test-User anlegen](#6-test-user-anlegen)
7. [Services konfigurieren](#7-services-konfigurieren)
8. [Token testen](#8-token-testen)
9. [Troubleshooting](#9-troubleshooting)

---

## 1. Docker Compose Setup

### docker-compose.yml erweitern

Füge Keycloak zu deiner `docker-compose.yml` hinzu:

```yaml
version: '3.8'

services:
  # PostgreSQL für alle Services + Keycloak
  postgres:
    image: postgres:16-alpine
    container_name: toxictoast-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: toxictoast
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    networks:
      - toxictoast-network

  # Keycloak
  keycloak:
    image: quay.io/keycloak/keycloak:23.0
    container_name: toxictoast-keycloak
    command: start-dev
    environment:
      KEYCLOAK_ADMIN: admin
      KEYCLOAK_ADMIN_PASSWORD: admin
      KC_DB: postgres
      KC_DB_URL: jdbc:postgresql://postgres:5432/toxictoast
      KC_DB_USERNAME: postgres
      KC_DB_PASSWORD: postgres
      KC_HOSTNAME: localhost
      KC_HTTP_PORT: 8080
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    networks:
      - toxictoast-network

  # Redpanda (Kafka)
  redpanda:
    image: docker.redpanda.com/vectorized/redpanda:latest
    container_name: toxictoast-redpanda
    command:
      - redpanda
      - start
      - --smp 1
      - --memory 1G
      - --reserve-memory 0M
      - --overprovisioned
      - --node-id 0
      - --kafka-addr PLAINTEXT://0.0.0.0:29092,OUTSIDE://0.0.0.0:9092
      - --advertise-kafka-addr PLAINTEXT://redpanda:29092,OUTSIDE://localhost:9092
    ports:
      - "9092:9092"
      - "9644:9644"
      - "29092:29092"
    networks:
      - toxictoast-network

volumes:
  postgres_data:

networks:
  toxictoast-network:
    driver: bridge
```

### Services starten

```bash
docker-compose up -d postgres keycloak
```

Warte ca. 30-60 Sekunden bis Keycloak vollständig gestartet ist.

---

## 2. Keycloak Admin Console

### Zugriff

1. Öffne Browser: **http://localhost:8080**
2. Klick auf **"Administration Console"**
3. Login:
   - **Username:** `admin`
   - **Password:** `admin`

---

## 3. Realm erstellen

Ein Realm ist eine isolierte Umgebung für deine User, Clients und Roles.

### Schritte:

1. **Realm erstellen:**
   - Oben links auf **"master"** klicken (Dropdown)
   - **"Create Realm"** wählen
   - **Realm name:** `toxictoast`
   - **Enabled:** `ON`
   - **Create** klicken

2. **Realm Settings konfigurieren:**
   - Links im Menü: **"Realm settings"**
   - Tab **"General":**
     - **Display name:** `ToxicToast Services`
     - **Enabled:** `ON`
   - Tab **"Login":**
     - **User registration:** `ON` (wenn du öffentliche Registrierung willst)
     - **Forgot password:** `ON`
     - **Remember me:** `ON`
     - **Email as username:** `ON` (optional, empfohlen)
   - Tab **"Email":**
     - Für Production: SMTP Server konfigurieren
     - Für Development: Kannst du überspringen
   - **Save** klicken

---

## 4. Clients konfigurieren

Clients = deine Services die mit Keycloak sprechen.

### 4.1 Blog Service Client

1. **Client erstellen:**
   - Links: **"Clients"** → **"Create client"**
   - **Client type:** `OpenID Connect`
   - **Client ID:** `blog-service`
   - **Next**

2. **Capability config:**
   - **Client authentication:** `ON`
   - **Authorization:** `OFF`
   - **Authentication flow:**
     - ✅ Standard flow
     - ✅ Direct access grants
     - ✅ Service accounts roles
   - **Next**

3. **Login settings:**
   - **Root URL:** `http://localhost:8081`
   - **Valid redirect URIs:** `*` (für Dev, später spezifischer)
   - **Web origins:** `*` (für CORS, später spezifischer)
   - **Save**

4. **Credentials holen:**
   - Tab **"Credentials"**
   - **Client secret:** Kopieren (brauchst du später)

### 4.2 Weitere Service Clients erstellen

Wiederhole die Schritte für alle Services:

| Service | Client ID | Port |
|---------|-----------|------|
| Blog Service | `blog-service` | 8081 |
| Link Service | `link-service` | 8082 |
| Foodfolio Service | `foodfolio-service` | 8083 |
| SSE Service | `sse-service` | 8084 |
| Webhook Service | `webhook-service` | 8085 |
| Notification Service | `notification-service` | 8086 |
| Twitchbot Service | `twitchbot-service` | 8087 |

### 4.3 Frontend Client (für Web/Mobile Apps)

Wenn du später ein Frontend hast:

1. **Client erstellen:**
   - **Client ID:** `toxictoast-frontend`
   - **Client type:** `OpenID Connect`
   - **Next**

2. **Capability config:**
   - **Client authentication:** `OFF` (Public Client)
   - **Authorization:** `OFF`
   - **Authentication flow:**
     - ✅ Standard flow
     - ✅ Direct access grants (für Login Form)
   - **Next**

3. **Login settings:**
   - **Root URL:** `http://localhost:3000` (z.B. React App)
   - **Valid redirect URIs:** `http://localhost:3000/*`
   - **Web origins:** `http://localhost:3000`
   - **Save**

---

## 5. Rollen (RBAC) einrichten

### 5.1 Realm Roles erstellen

Realm Roles = globale Rollen über alle Clients hinweg.

1. **Roles erstellen:**
   - Links: **"Realm roles"** → **"Create role"**

2. **Standard Roles:**

   **Admin Role:**
   - **Role name:** `admin`
   - **Description:** `Administrator with full access`
   - **Save**

   **User Role:**
   - **Role name:** `user`
   - **Description:** `Standard user`
   - **Save**

   **Moderator Role:**
   - **Role name:** `moderator`
   - **Description:** `Content moderator`
   - **Save**

   **Premium User Role:**
   - **Role name:** `premium`
   - **Description:** `Premium subscriber`
   - **Save**

### 5.2 Client-spezifische Roles (optional)

Für fein-granulare Permissions pro Service:

1. **Zum Client gehen:**
   - Links: **"Clients"** → z.B. `blog-service`
   - Tab **"Roles"**
   - **Create role**

2. **Beispiel Blog Service Roles:**
   - `blog:post:create`
   - `blog:post:edit`
   - `blog:post:delete`
   - `blog:post:publish`
   - `blog:comment:moderate`

### 5.3 Composite Roles (optional)

Kombiniere mehrere Roles:

1. **Admin = alle Permissions:**
   - **"Realm roles"** → `admin` (Edit)
   - Tab **"Associated roles"**
   - **Assign role** → Wähle `user`, `moderator`, `premium`
   - **Assign**

---

## 6. Test-User anlegen

### 6.1 Admin User erstellen

1. **User erstellen:**
   - Links: **"Users"** → **"Add user"**
   - **Username:** `admin@toxictoast.de`
   - **Email:** `admin@toxictoast.de`
   - **First name:** `Admin`
   - **Last name:** `User`
   - **Email verified:** `ON`
   - **Enabled:** `ON`
   - **Create**

2. **Password setzen:**
   - Tab **"Credentials"**
   - **Set password:**
     - **Password:** `admin123`
     - **Password confirmation:** `admin123`
     - **Temporary:** `OFF` (sonst muss User beim ersten Login ändern)
   - **Save**

3. **Roles zuweisen:**
   - Tab **"Role mapping"**
   - **Assign role** → Filter: `Filter by realm roles`
   - Wähle: `admin`
   - **Assign**

### 6.2 Standard User erstellen

1. **User erstellen:**
   - Username: `user@toxictoast.de`
   - Email: `user@toxictoast.de`
   - First name: `Test`
   - Last name: `User`
   - Email verified: `ON`
   - **Create**

2. **Password:** `user123` (Temporary: OFF)

3. **Roles:** `user`

### 6.3 Moderator User erstellen

- Username: `mod@toxictoast.de`
- Password: `mod123`
- Roles: `user`, `moderator`

---

## 7. Services konfigurieren

### 7.1 .env Datei erstellen

Erstelle eine `.env` im Root-Verzeichnis:

```env
# Environment
ENVIRONMENT=development

# Database (Shared)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# Kafka
KAFKA_BROKERS=localhost:9092

# Keycloak (Shared)
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=toxictoast
AUTH_ENABLED=true

# Blog Service
BLOG_PORT=8081
BLOG_GRPC_PORT=9091
BLOG_DB_NAME=toxictoast
KEYCLOAK_CLIENT_ID_BLOG=blog-service
KEYCLOAK_CLIENT_SECRET_BLOG=<dein-client-secret>

# Link Service
LINK_PORT=8082
LINK_GRPC_PORT=9092
LINK_DB_NAME=toxictoast
KEYCLOAK_CLIENT_ID_LINK=link-service
KEYCLOAK_CLIENT_SECRET_LINK=<dein-client-secret>

# Foodfolio Service
FOODFOLIO_PORT=8083
FOODFOLIO_GRPC_PORT=9093
FOODFOLIO_DB_NAME=toxictoast
KEYCLOAK_CLIENT_ID_FOODFOLIO=foodfolio-service
KEYCLOAK_CLIENT_SECRET_FOODFOLIO=<dein-client-secret>

# SSE Service
SSE_PORT=8084
SSE_GRPC_PORT=9094
KEYCLOAK_CLIENT_ID_SSE=sse-service

# Webhook Service
WEBHOOK_PORT=8085
WEBHOOK_GRPC_PORT=9095
WEBHOOK_DB_NAME=toxictoast
KEYCLOAK_CLIENT_ID_WEBHOOK=webhook-service

# Notification Service
NOTIFICATION_PORT=8086
NOTIFICATION_GRPC_PORT=9096
NOTIFICATION_DB_NAME=toxictoast
KEYCLOAK_CLIENT_ID_NOTIFICATION=notification-service

# Twitchbot Service
TWITCHBOT_PORT=8087
TWITCHBOT_GRPC_PORT=9097
TWITCHBOT_DB_NAME=toxictoast
KEYCLOAK_CLIENT_ID_TWITCHBOT=twitchbot-service
KEYCLOAK_CLIENT_SECRET_TWITCHBOT=<dein-client-secret>
```

### 7.2 Client Secrets eintragen

Für jeden Service:

1. Keycloak öffnen
2. **Clients** → z.B. `blog-service`
3. Tab **"Credentials"**
4. **Client secret** kopieren
5. In `.env` eintragen bei `KEYCLOAK_CLIENT_SECRET_BLOG`

### 7.3 Services Config updaten

Deine Services sollten bereits die Env-Variablen aus `.env` lesen (godotenv).

**Beispiel: blog-service/pkg/config/config.go**

```go
Keycloak: sharedConfig.KeycloakConfig{
    URL:          getEnv("KEYCLOAK_URL", "http://localhost:8080"),
    Realm:        getEnv("KEYCLOAK_REALM", "toxictoast"),
    ClientID:     getEnv("KEYCLOAK_CLIENT_ID_BLOG", "blog-service"),
    ClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET_BLOG", ""),
},
AuthEnabled: getEnvAsBool("AUTH_ENABLED", false),
```

---

## 8. Token testen

### 8.1 Token mit curl holen

```bash
# Admin User Token
curl -X POST "http://localhost:8080/realms/toxictoast/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin@toxictoast.de" \
  -d "password=admin123" \
  -d "grant_type=password" \
  -d "client_id=blog-service" \
  -d "client_secret=<dein-blog-service-secret>"
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cC...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cC...",
  "token_type": "Bearer"
}
```

### 8.2 Token decodieren

Kopiere den `access_token` und gehe zu: **https://jwt.io**

Du solltest sehen:
```json
{
  "realm_access": {
    "roles": ["admin", "user"]
  },
  "email": "admin@toxictoast.de",
  "preferred_username": "admin@toxictoast.de"
}
```

### 8.3 Service mit Token aufrufen

```bash
# Token in Variable speichern
export TOKEN="eyJhbGciOiJSUzI1NiIsInR5cC..."

# Blog Service API Call (wenn du einen Endpoint hast)
curl -X GET "http://localhost:8081/health" \
  -H "Authorization: Bearer $TOKEN"
```

### 8.4 Mit grpcurl testen

```bash
# Token als Metadata mitgeben
grpcurl -plaintext \
  -H "Authorization: Bearer $TOKEN" \
  -d '{}' \
  localhost:9091 \
  blog.BlogService/ListPosts
```

---

## 9. Troubleshooting

### Problem: "User not found"

**Lösung:**
- Prüfe ob User in Keycloak existiert (Users → Search)
- Prüfe ob Email verified ist
- Prüfe ob User enabled ist

### Problem: "Invalid credentials"

**Lösung:**
- Prüfe Username/Email (ist Email as username enabled?)
- Prüfe Password
- Prüfe ob Temporary Password abgelaufen ist

### Problem: "Client not found"

**Lösung:**
- Prüfe `KEYCLOAK_REALM` in .env
- Prüfe `KEYCLOAK_CLIENT_ID` in .env
- Prüfe ob Client im richtigen Realm ist

### Problem: "Invalid token"

**Lösung:**
```bash
# Token Validierung testen
curl "http://localhost:8080/realms/toxictoast/protocol/openid-connect/userinfo" \
  -H "Authorization: Bearer $TOKEN"
```

- Prüfe ob Token abgelaufen ist (expires_in)
- Prüfe ob KEYCLOAK_URL erreichbar ist
- Prüfe ob Realm Name korrekt ist

### Problem: "Insufficient permissions"

**Lösung:**
- User hat nicht die richtige Role
- Keycloak → Users → [User] → Role mapping
- Rolle zuweisen und neu einloggen (neues Token holen)

### Problem: Services können Keycloak nicht erreichen

**Docker Compose:**
- Services müssen `keycloak:8080` nutzen (nicht `localhost:8080`)
- In `.env` für Docker: `KEYCLOAK_URL=http://keycloak:8080`

**Local Development:**
- Services nutzen `localhost:8080`
- In `.env`: `KEYCLOAK_URL=http://localhost:8080`

### Keycloak Logs ansehen

```bash
docker logs toxictoast-keycloak

# Live logs
docker logs -f toxictoast-keycloak
```

---

## Nächste Schritte

### Production Checklist:

- [ ] **SMTP Email konfigurieren** (Realm settings → Email)
- [ ] **HTTPS aktivieren** (mit Let's Encrypt / Reverse Proxy)
- [ ] **Admin Password ändern**
- [ ] **Redirect URIs spezifizieren** (nicht `*`)
- [ ] **CORS Web Origins spezifizieren** (nicht `*`)
- [ ] **Token Lifetimes anpassen** (Realm settings → Tokens)
- [ ] **Password Policy setzen** (Realm settings → Authentication → Policies)
- [ ] **Bruteforce Protection** (Realm settings → Security defenses)
- [ ] **Backup Strategy** für Keycloak DB
- [ ] **Client Secrets in Secrets Manager** (nicht in .env committen)

### Erweiterte Features:

- **Social Login:** Google, GitHub, etc. (Identity providers)
- **2FA/MFA:** TOTP einrichten (Authentication → Required actions)
- **Custom Themes:** Keycloak UI anpassen
- **User Federation:** LDAP/Active Directory
- **Events & Auditing:** Realm settings → Events

---

## Zusammenfassung

✅ Keycloak läuft in Docker
✅ Realm `toxictoast` erstellt
✅ Clients für alle Services konfiguriert
✅ Roles (admin, user, moderator, premium) angelegt
✅ Test-User mit verschiedenen Rollen erstellt
✅ Services via .env konfiguriert
✅ Token-Flow getestet

**Deine Services sind jetzt gesichert mit Keycloak!**

Bei Fragen: Keycloak Dokumentation → https://www.keycloak.org/documentation
