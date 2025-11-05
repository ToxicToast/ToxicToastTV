# Foodfolio Service - Quick Start Guide

## 🚀 Schnellstart

### Option 1: Lokale Entwicklung (mit Docker Dependencies)

**1. Dependencies starten (PostgreSQL + Kafka)**
```bash
make docker-dev
```

**2. Service lokal ausführen**
```bash
make run
```

Der Service ist nun erreichbar unter:
- HTTP Health Check: http://localhost:8889/health
- gRPC: localhost:9091

---

### Option 2: Alles mit Docker Compose

**1. Alle Services starten**
```bash
make docker-up
```

**2. Logs anschauen**
```bash
make docker-logs
```

**3. Services stoppen**
```bash
make docker-down
```

---

## 📝 Umgebungsvariablen

Kopiere `.env.example` zu `.env` und passe die Werte an:
```bash
cp .env.example .env
```

Wichtige Variablen:
- `DB_HOST`: PostgreSQL Host (default: localhost)
- `DB_PORT`: PostgreSQL Port (default: 5432)
- `GRPC_PORT`: gRPC Server Port (default: 9091)
- `PORT`: HTTP Server Port (default: 8889)
- `KAFKA_BROKERS`: Kafka Broker Address (default: localhost:19092)

---

## 🧪 Service testen

### Health Check
```bash
curl http://localhost:8889/health
```

### gRPC Reflection (mit grpcurl)
```bash
# Installiere grpcurl (falls nicht vorhanden)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Liste alle Services auf
grpcurl -plaintext localhost:9091 list

# Liste alle Methods eines Service
grpcurl -plaintext localhost:9091 list foodfolio.v1.ItemService

# Erstelle eine Company
grpcurl -plaintext -d '{"name": "Coca Cola"}' \
  localhost:9091 foodfolio.v1.CompanyService/CreateCompany

# Liste alle Companies
grpcurl -plaintext -d '{"pagination": {"page": 1, "page_size": 10}}' \
  localhost:9091 foodfolio.v1.CompanyService/ListCompanies
```

---

## 📊 Verfügbare Services

Der Microservice bietet folgende gRPC Services:

1. **CategoryService** - Produkt-Kategorien (hierarchisch)
2. **CompanyService** - Hersteller/Marken
3. **TypeService** - Verpackungsarten (Dose, PET, Box, etc.)
4. **SizeService** - Größenangaben
5. **WarehouseService** - Einkaufsstandorte
6. **LocationService** - Lagerorte (hierarchisch)
7. **ItemService** - Produkte
8. **ItemVariantService** - Produktvarianten mit Stock-Tracking
9. **ItemDetailService** - Einzelne physische Items mit MHD
10. **ShoppinglistService** - Einkaufslisten mit Auto-Generation
11. **ReceiptService** - Kassenbons mit OCR

Jeder Service bietet Standard CRUD-Operationen plus spezialisierte Funktionen.

---

## 🔧 Entwicklung

### Proto Files neu generieren
```bash
make proto
```

### Service bauen
```bash
make build
```

### Tests ausführen
```bash
make test
```

### Test Coverage
```bash
make test-coverage
```

---

## 🐳 Docker Commands

```bash
# Docker Image bauen
make docker-build

# Alle Services starten
make docker-up

# Nur Dependencies starten
make docker-dev

# Logs anzeigen
make docker-logs

# Services neustarten
make docker-restart

# Services stoppen
make docker-down

# Aufräumen (inkl. Volumes)
make docker-clean
```

---

## 🌐 Redpanda Console

Wenn Kafka/Redpanda läuft, ist die Web UI verfügbar:
**http://localhost:8080**

Hier kannst du:
- Topics ansehen
- Messages publishen/konsumieren
- Consumer Groups monitoren

---

## 📚 Beispiel-Workflow

### 1. Grunddaten anlegen
```bash
# Company erstellen
grpcurl -plaintext -d '{"name": "Coca Cola"}' \
  localhost:9091 foodfolio.v1.CompanyService/CreateCompany

# Type erstellen
grpcurl -plaintext -d '{"name": "Dose"}' \
  localhost:9091 foodfolio.v1.TypeService/CreateType

# Size erstellen
grpcurl -plaintext -d '{"name": "Standard", "value": 0.33, "unit": "L"}' \
  localhost:9091 foodfolio.v1.SizeService/CreateSize

# Category erstellen
grpcurl -plaintext -d '{"name": "Getränke"}' \
  localhost:9091 foodfolio.v1.CategoryService/CreateCategory
```

### 2. Item erstellen
```bash
grpcurl -plaintext -d '{
  "name": "Coca Cola",
  "category_id": "<category-uuid>",
  "company_id": "<company-uuid>",
  "type_id": "<type-uuid>"
}' localhost:9091 foodfolio.v1.ItemService/CreateItem
```

### 3. ItemVariant mit Stock Tracking
```bash
grpcurl -plaintext -d '{
  "item_id": "<item-uuid>",
  "size_id": "<size-uuid>",
  "variant_name": "Standard",
  "min_sku": 10,
  "max_sku": 50,
  "is_normally_frozen": false
}' localhost:9091 foodfolio.v1.ItemVariantService/CreateItemVariant
```

### 4. Low Stock Check
```bash
grpcurl -plaintext -d '{"pagination": {"page": 1, "page_size": 10}}' \
  localhost:9091 foodfolio.v1.ItemVariantService/GetLowStockVariants
```

### 5. Shopping List Auto-Generate
```bash
grpcurl -plaintext -d '{"name": "Weekly Shopping"}' \
  localhost:9091 foodfolio.v1.ShoppinglistService/GenerateFromLowStock
```

---

## 🛠 Troubleshooting

### Service startet nicht
```bash
# Logs prüfen
make docker-logs

# Einzelne Container prüfen
docker ps
docker logs foodfolio-service
docker logs foodfolio-postgres
docker logs foodfolio-redpanda
```

### Database Connection Error
```bash
# Prüfe ob PostgreSQL läuft
docker ps | grep postgres

# PostgreSQL Logs
docker logs foodfolio-postgres

# In die Database connecten
docker exec -it foodfolio-postgres psql -U foodfolio -d foodfolio
```

### Proto Generation Fehler
```bash
# Protoc installieren (falls nicht vorhanden)
# Siehe: https://grpc.io/docs/protoc-installation/

# Proto Files neu generieren
make proto-clean
make proto-gen
```

---

## 📖 Weitere Dokumentation

- [README.md](README.md) - Vollständige Dokumentation
- [API Proto Files](api/proto/) - gRPC Service Definitionen
- [Architecture](docs/ARCHITECTURE.md) - Service Architecture (TODO)

---

## 🤝 Support

Bei Fragen oder Problemen:
1. Prüfe die Logs: `make docker-logs`
2. Checke Health Endpoint: `curl http://localhost:8889/health`
3. Verifiziere .env Konfiguration
4. Prüfe Docker Container Status: `docker ps`
