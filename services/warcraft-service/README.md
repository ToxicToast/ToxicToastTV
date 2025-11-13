# Warcraft Service

World of Warcraft character and guild tracking service with Blizzard API integration.

## Features

### Character Management
- Track multiple WoW characters across regions
- Character details (level, item level, class, race, faction)
- Equipment tracking
- Character stats monitoring
- Automatic data refresh from Blizzard API

### Guild Management
- Track guild information
- Guild roster management
- Achievement tracking
- Member statistics

### Reference Data
- Normalized character races
- Character classes
- Factions (Alliance/Horde)

## Architecture

```
warcraft-service/
├── api/proto/              # gRPC Protocol Buffer definitions
├── cmd/server/             # Service entry point
├── internal/
│   ├── domain/            # Domain entities (Character, Guild, Race, etc.)
│   ├── repository/        # Data access layer
│   │   ├── entity/       # Database entities
│   │   ├── mapper/       # Domain ↔ Entity mapping
│   │   └── impl/         # Repository implementations
│   ├── usecase/          # Business logic
│   └── handler/grpc/     # gRPC handlers
├── pkg/
│   ├── config/           # Configuration management
│   └── blizzard/         # Blizzard API client
└── migrations/           # Database migrations (if needed)
```

## Data Model

### Character
- **Character**: Base entity (name, realm, region for API calls)
- **CharacterDetails**: Full details (display_name, level, class, race, faction)
- **CharacterEquipment**: Equipped items per slot
- **CharacterStats**: Character stats (health, strength, etc.)

**Relationships:**
- 1 Character = 1 CharacterDetails
- 1 Character = 1 CharacterEquipment
- CharacterDetails → Race (FK)
- CharacterDetails → Class (FK)
- CharacterDetails → Faction (FK)

### Guild
- **Guild**: Guild entity with faction FK
- **GuildMember**: Guild roster members

### Reference Data
- **Race**: Playable races (Human, Orc, etc.)
- **Class**: Playable classes (Warrior, Mage, etc.)
- **Faction**: Alliance or Horde

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=warcraft_db

# Blizzard API Credentials
# Get from: https://develop.battle.net/access/clients
BLIZZARD_CLIENT_ID=your_client_id
BLIZZARD_CLIENT_SECRET=your_client_secret
BLIZZARD_REGION=us
```

## Getting Blizzard API Credentials

1. Go to https://develop.battle.net/access/clients
2. Create a new client
3. Copy Client ID and Client Secret
4. Add to your `.env` file

## Running the Service

### Local Development

```bash
# Install dependencies
go mod download

# Run service
go run cmd/server/main.go
```

### Docker

```bash
# Build from repository root
docker build -f services/warcraft-service/Dockerfile -t warcraft-service .

# Run
docker run -p 9090:9090 \
  -e DB_HOST=postgres \
  -e DB_NAME=warcraft_db \
  -e BLIZZARD_CLIENT_ID=your_id \
  -e BLIZZARD_CLIENT_SECRET=your_secret \
  warcraft-service
```

## gRPC API

### CharacterService

```protobuf
service CharacterService {
  rpc CreateCharacter(CreateCharacterRequest) returns (CharacterResponse);
  rpc GetCharacter(GetCharacterRequest) returns (CharacterResponse);
  rpc ListCharacters(ListCharactersRequest) returns (ListCharactersResponse);
  rpc UpdateCharacter(UpdateCharacterRequest) returns (CharacterResponse);
  rpc DeleteCharacter(DeleteCharacterRequest) returns (DeleteCharacterResponse);
  rpc RefreshCharacter(RefreshCharacterRequest) returns (CharacterResponse);
  rpc GetCharacterEquipment(GetCharacterEquipmentRequest) returns (CharacterEquipmentResponse);
  rpc GetCharacterStats(GetCharacterStatsRequest) returns (CharacterStatsResponse);
}
```

### GuildService

```protobuf
service GuildService {
  rpc CreateGuild(CreateGuildRequest) returns (GuildResponse);
  rpc GetGuild(GetGuildRequest) returns (GuildResponse);
  rpc ListGuilds(ListGuildsRequest) returns (ListGuildsResponse);
  rpc UpdateGuild(UpdateGuildRequest) returns (GuildResponse);
  rpc DeleteGuild(DeleteGuildRequest) returns (DeleteGuildResponse);
  rpc RefreshGuild(RefreshGuildRequest) returns (GuildResponse);
  rpc GetGuildRoster(GetGuildRosterRequest) returns (GuildRosterResponse);
}
```

## Usage Examples

### Using grpcurl

```bash
# Create character
grpcurl -plaintext -d '{
  "name": "ragnaros",
  "realm": "Area 52",
  "region": "us"
}' localhost:9090 warcraft.CharacterService/CreateCharacter

# List characters
grpcurl -plaintext -d '{
  "page": 1,
  "page_size": 10
}' localhost:9090 warcraft.CharacterService/ListCharacters

# Refresh character data from Blizzard API
grpcurl -plaintext -d '{
  "id": "character-uuid"
}' localhost:9090 warcraft.CharacterService/RefreshCharacter
```

## Blizzard API Integration

### Current Status: Stub Implementation

The Blizzard API client (`pkg/blizzard/client.go`) is currently a stub. To implement full integration:

1. **OAuth2 Authentication**
   - Get access token using client credentials
   - Token endpoint: `https://oauth.battle.net/token`

2. **Character Profile API**
   - Endpoint: `GET https://{region}.api.blizzard.com/profile/wow/character/{realm}/{name}`
   - Namespace: `profile-{region}`

3. **Guild API**
   - Endpoint: `GET https://{region}.api.blizzard.com/data/wow/guild/{realm}/{name}`
   - Namespace: `profile-{region}`

### API Documentation
- Official Docs: https://develop.battle.net/documentation/world-of-warcraft
- Game Data APIs: https://develop.battle.net/documentation/world-of-warcraft/game-data-apis

## Database Schema

Auto-migrated via GORM:

```sql
-- Factions
CREATE TABLE factions (
  id UUID PRIMARY KEY,
  key VARCHAR UNIQUE NOT NULL,
  name VARCHAR NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Races
CREATE TABLE races (
  id UUID PRIMARY KEY,
  key VARCHAR UNIQUE NOT NULL,
  name VARCHAR NOT NULL,
  faction_id UUID REFERENCES factions(id),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Classes
CREATE TABLE classes (
  id UUID PRIMARY KEY,
  key VARCHAR UNIQUE NOT NULL,
  name VARCHAR NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Characters
CREATE TABLE characters (
  id UUID PRIMARY KEY,
  name VARCHAR NOT NULL,
  realm VARCHAR NOT NULL,
  region VARCHAR NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE(name, realm, region)
);

-- Character Details
CREATE TABLE character_details (
  id UUID PRIMARY KEY,
  character_id UUID UNIQUE REFERENCES characters(id),
  display_name VARCHAR NOT NULL,
  display_realm VARCHAR NOT NULL,
  level INT NOT NULL,
  item_level INT NOT NULL,
  class_id UUID REFERENCES classes(id),
  race_id UUID REFERENCES races(id),
  faction_id UUID REFERENCES factions(id),
  guild_id UUID REFERENCES guilds(id),
  thumbnail_url TEXT,
  achievement_points INT DEFAULT 0,
  last_synced_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Guilds
CREATE TABLE guilds (
  id UUID PRIMARY KEY,
  name VARCHAR NOT NULL,
  realm VARCHAR NOT NULL,
  region VARCHAR NOT NULL,
  faction_id UUID REFERENCES factions(id),
  member_count INT DEFAULT 0,
  achievement_points INT DEFAULT 0,
  last_synced_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE(name, realm, region)
);
```

## Development

### Generate Proto Files

```bash
cd services/warcraft-service
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/warcraft.proto
```

### Build

```bash
go build ./...
```

### Test

```bash
go test ./...
```

## TODO

- [ ] Implement full Blizzard API integration
- [ ] Add OAuth2 token caching
- [ ] Implement equipment endpoint
- [ ] Implement stats endpoint
- [ ] Add guild roster sync
- [ ] Add background job for auto-refresh
- [ ] Add rate limiting for API calls
- [ ] Add Kafka event publishing
- [ ] Add integration tests

## Tech Stack

- **Go 1.24+**
- **gRPC** - Service communication
- **PostgreSQL + GORM** - Data persistence
- **Protocol Buffers** - API definition
- **Blizzard Battle.net API** - Data source

## License

Proprietary - ToxicToast
