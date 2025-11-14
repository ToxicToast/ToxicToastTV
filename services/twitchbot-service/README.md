# Twitchbot Service

A Twitch bot service for tracking and managing stream data, built with Go, gRPC, and Clean Architecture.

## Features

- **Stream Management** - Track stream sessions with viewer and message statistics
- **Message Logging** - Log and analyze chat messages with full-text search
  - **24/7 Message Logging** - All messages logged even when offline (Chat-Only stream)
- **Multi-Channel Support** - Bot can join and monitor multiple Twitch channels simultaneously
- **Bot Management** - Control bot via gRPC (join/leave channels, send messages, get status)
- **Viewer Tracking** - Monitor viewer engagement and participation
  - **Per-Channel Viewer Lists** - Track which viewers are in which channels independently
  - **Automatic Viewer Fetching** - When joining a channel, automatically fetch all current viewers
- **Clip Archival** - Archive and manage stream clips
- **Command System** - Custom chat commands with permissions and cooldowns
- **gRPC API** - High-performance API for all service operations (7 services, 43 endpoints)
- **Event Publishing** - Kafka events for stream, message, viewer, clip, and command activities
- **Auto Token Refresh** - Automatic Twitch token refresh with reconnect
- **Clean Architecture** - Maintainable and testable codebase
- **Background Jobs** - Automatic cleanup of old messages and inactive stream sessions

### Background Jobs

The service includes two automated background job schedulers:

1. **Message Cleanup Scheduler** (default: every 24 hours)
   - Permanently deletes chat messages older than 90 days
   - Helps maintain database performance
   - Retention period configurable via environment variables

2. **Stream Session Closer** (default: every hour)
   - Automatically closes stream sessions inactive for 24+ hours
   - Prevents orphaned active streams
   - Handles edge cases where EndStream wasn't called

## Architecture

This service follows Clean Architecture principles:

```
twitchbot-service/
├── api/proto/              # gRPC proto definitions
├── cmd/server/             # Service entry point
├── internal/
│   ├── domain/             # Domain models (Stream, Message, Viewer, Clip, Command)
│   ├── repository/         # Data access layer (PostgreSQL/GORM)
│   ├── usecase/            # Business logic layer
│   └── handler/grpc/       # gRPC handlers
├── pkg/
│   └── config/             # Service configuration
└── migrations/             # Database migrations
```

## Tech Stack

- **Go 1.24+**
- **gRPC** with Protocol Buffers
- **PostgreSQL** with GORM
- **Kafka/Redpanda** for event publishing
- **Keycloak** for authentication (optional)

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 14+
- Kafka/Redpanda (optional)
- **Twitch Credentials** (optional - see [TWITCH_SETUP.md](./TWITCH_SETUP.md))

**Note**: Service works without Twitch credentials in **API-only mode**!

### Installation

1. **Clone the repository**
   ```bash
   cd services/twitchbot-service
   ```

2. **Copy environment file**
   ```bash
   cp .env.example .env
   ```

3. **Configure Twitch credentials (OPTIONAL)**

   **Option A**: Skip this step to run in API-only mode (no bot, only gRPC endpoints)

   **Option B**: Enable bot by editing `.env` with your credentials:
   ```env
   TWITCH_CHANNEL=your_channel_name
   TWITCH_ACCESS_TOKEN=oauth:your_token_here
   TWITCH_BOT_USERNAME=your_bot_username
   ```

   📖 **[Full Twitch Setup Guide →](./TWITCH_SETUP.md)**

4. **Set up database**
   ```bash
   createdb twitchbot_db
   ```

5. **Install dependencies**
   ```bash
   go mod download
   ```

6. **Generate proto files**
   ```bash
   protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. api/proto/twitchbot.proto
   ```

7. **Run the service**
   ```bash
   go run cmd/server/main.go
   ```

The service will start:
- gRPC server on port `9093`
- HTTP health checks on port `8083`

## API Services

### StreamService
- `CreateStream` - Start a new stream session
- `GetStream` - Get stream by ID
- `ListStreams` - List streams with filtering
- `UpdateStream` - Update stream metadata
- `EndStream` - End an active stream
- `GetActiveStream` - Get currently active stream
- `GetStreamStats` - Get stream statistics

### MessageService
- `CreateMessage` - Log a chat message
- `GetMessage` - Get message by ID
- `ListMessages` - List messages with filtering
- `SearchMessages` - Full-text search messages
- `GetMessageStats` - Get message statistics
- `DeleteMessage` - Soft delete a message

### ViewerService
- `CreateViewer` - Register a new viewer
- `GetViewer` - Get viewer by ID
- `GetViewerByTwitchId` - Get viewer by Twitch ID
- `ListViewers` - List viewers with sorting
- `UpdateViewer` - Update viewer information
- `GetViewerStats` - Get viewer statistics
- `DeleteViewer` - Soft delete a viewer

### ClipService
- `CreateClip` - Archive a new clip
- `GetClip` - Get clip by ID
- `GetClipByTwitchId` - Get clip by Twitch clip ID
- `ListClips` - List clips with sorting
- `UpdateClip` - Update clip metadata
- `DeleteClip` - Soft delete a clip

### CommandService
- `CreateCommand` - Create a new command
- `GetCommand` - Get command by ID
- `GetCommandByName` - Get command by name
- `ListCommands` - List commands
- `UpdateCommand` - Update command configuration
- `ExecuteCommand` - Execute a command with permission checks
- `DeleteCommand` - Soft delete a command

### ChannelViewerService
- `GetChannelViewer` - Get a specific viewer in a channel
- `ListChannelViewers` - List all viewers in a channel (with pagination)
- `CountChannelViewers` - Count total viewers in a channel
- `RemoveChannelViewer` - Remove a viewer from channel tracking

## Multi-Channel Stream Tracking

The bot can join multiple channels and **tracks each channel's stream independently**!

### Per-Channel Stream Management

Each channel the bot joins has its own stream tracking:

1. **Bot joins channel** → Checks if channel is live via Helix API
2. **Fetches all current viewers** → Adds them to the database (requires `moderator:read:chatters` scope)
3. **Channel is live?** → Creates dedicated stream for that channel
4. **Channel offline?** → Uses shared Chat-Only stream (`00000000-0000-0000-0000-000000000001`)
5. **Channel goes live** → Automatically switches to new stream
6. **Stream ends** → Switches back to Chat-Only stream

### Example: Multi-Channel Setup

```bash
# Bot starts in your channel (toxictoast)
# - toxictoast offline → uses Chat-Only stream
# - toxictoast goes live → creates Stream A

# Join another channel
grpcurl -plaintext -d '{"channel": "shroud"}' \
  localhost:9093 twitchbot.BotService/JoinChannel

# Response: Bot joins channel, fetches all current viewers, checks stream status
# - shroud is live → creates Stream B immediately
# - shroud messages → go to Stream B
# - toxictoast messages → still go to Stream A
# - Both channels maintain independent viewer lists
```

### Stream & Viewer Isolation

- ✅ Each channel gets its own stream when live
- ✅ Messages are logged to the correct stream
- ✅ Stream stats (viewers, messages) are tracked per channel
- ✅ **Viewers are tracked per channel** - same user can be in multiple channels
- ✅ When joining a channel, all current viewers are automatically fetched
- ✅ Viewer list updates in real-time as users send messages
- ✅ When offline, all channels share the Chat-Only stream
- ✅ Stream watcher polls ALL joined channels every 30 seconds

This means you can monitor multiple streamers and each gets their own accurate stream data and viewer list!

## Usage Examples

### Bot Management

```bash
# Join another channel
grpcurl -plaintext -d '{"channel": "shroud"}' \
  localhost:9093 twitchbot.BotService/JoinChannel

# Leave a channel
grpcurl -plaintext -d '{"channel": "shroud"}' \
  localhost:9093 twitchbot.BotService/LeaveChannel

# List all joined channels
grpcurl -plaintext localhost:9093 twitchbot.BotService/ListChannels

# Get bot status
grpcurl -plaintext localhost:9093 twitchbot.BotService/GetBotStatus

# Send a message
grpcurl -plaintext -d '{"channel": "toxictoast", "message": "Hello!"}' \
  localhost:9093 twitchbot.BotService/SendMessage
```

### Stream & Message Queries

```bash
# Get active stream
grpcurl -plaintext localhost:9093 twitchbot.StreamService/GetActiveStream

# List recent messages (includes Chat-Only messages!)
grpcurl -plaintext -d '{"stream_id": "00000000-0000-0000-0000-000000000001", "limit": 10}' \
  localhost:9093 twitchbot.MessageService/ListMessages

# Execute a command
grpcurl -plaintext -d '{
  "command_name": "!hello",
  "user_id": "12345",
  "username": "viewer",
  "is_moderator": false,
  "is_subscriber": true
}' localhost:9093 twitchbot.CommandService/ExecuteCommand
```

### Channel Viewer Queries

```bash
# List viewers in a channel
grpcurl -plaintext -d '{"channel": "toxictoast", "limit": 50, "offset": 0}' \
  localhost:9093 twitchbot.ChannelViewerService/ListChannelViewers

# Count viewers in a channel
grpcurl -plaintext -d '{"channel": "toxictoast"}' \
  localhost:9093 twitchbot.ChannelViewerService/CountChannelViewers

# Get specific viewer in a channel
grpcurl -plaintext -d '{"channel": "toxictoast", "twitch_id": "12345"}' \
  localhost:9093 twitchbot.ChannelViewerService/GetChannelViewer
```

## Health Checks

- `GET /health` - General health check
- `GET /health/ready` - Readiness check (includes DB)
- `GET /health/live` - Liveness check

## Development

### Build
```bash
go build ./...
```

### Test
```bash
go test ./...
```

### Regenerate Proto Files
```bash
protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. api/proto/twitchbot.proto
```

## Environment Variables

See `.env.example` for all available configuration options.

Key variables:
- `TWITCH_CHANNEL` - Your Twitch channel name (required for bot mode)
- `TWITCH_ACCESS_TOKEN` - OAuth token for API access (required for bot mode)
- `TWITCH_BOT_USERNAME` - Bot account username (required for bot mode)
- `TWITCH_CLIENT_ID` - Twitch application client ID (optional, enables auto-refresh)
- `TWITCH_CLIENT_SECRET` - Twitch application secret (optional, enables auto-refresh)
- `DB_*` - Database connection settings
- `KAFKA_*` - Kafka/event streaming settings

**Automatic Token Refresh**: If you provide `TWITCH_CLIENT_ID` and `TWITCH_CLIENT_SECRET`, the service will automatically refresh your access token before it expires and reconnect as needed. See [TWITCH_SETUP.md](./TWITCH_SETUP.md) for details.

**Required Scopes for Full Functionality**:
- `chat:read` - Read chat messages (required)
- `chat:edit` - Send chat messages (required)
- `moderator:read:chatters` - Fetch viewer list when joining channels (optional, but recommended for per-channel viewer tracking)

Without `moderator:read:chatters`, the bot will still track viewers as they send messages, but won't fetch the initial viewer list when joining a channel.

## License

Proprietary - ToxicToast
