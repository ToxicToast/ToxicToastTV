# Database Migrations

This directory contains SQL migration files for the twitchbot-service database.

## Migration Files

### 001_initial_schema.sql
Creates all initial tables:
- `streams` - Stream sessions
- `messages` - Chat messages
- `viewers` - Viewer tracking
- `clips` - Clip archival
- `commands` - Custom commands

### 002_chat_only_stream.sql
Creates a permanent "Chat-Only" stream for logging messages when not live:
- **Fixed UUID**: `00000000-0000-0000-0000-000000000001`
- **Purpose**: All chat messages when stream is offline go to this stream
- **Title**: "Chat-Only Messages"
- **Game**: "Just Chatting"
- **Always Active**: This stream never ends

## Running Migrations

**Note**: The service uses GORM AutoMigrate, which automatically creates tables based on domain models. These SQL files are for reference and manual execution if needed.

### Manual Migration

```bash
# Connect to your PostgreSQL database
psql -U postgres -d twitchbot_db

# Run migrations in order
\i migrations/001_initial_schema.sql
\i migrations/002_chat_only_stream.sql
```

### Verify Chat-Only Stream

```sql
SELECT id, title, game_name, is_active, started_at
FROM streams
WHERE id = '00000000-0000-0000-0000-000000000001';
```

Expected output:
```
                  id                  |       title        |  game_name   | is_active |         started_at
--------------------------------------+--------------------+--------------+-----------+----------------------------
 00000000-0000-0000-0000-000000000001 | Chat-Only Messages | Just Chatting| t         | 2025-11-06 10:00:00.000000
```

## How Chat-Only Stream Works

1. **Service starts** → Bot connects to IRC
2. **Message arrives** → Bot checks for active stream
3. **No live stream?** → Bot uses Chat-Only stream (UUID `00000000-0000-0000-0000-000000000001`)
4. **Stream goes live** → Bot switches to real stream
5. **Stream ends** → Bot switches back to Chat-Only stream

This ensures **ALL** chat messages are logged, regardless of stream status!

## Notes

- The Chat-Only stream is created with `ON CONFLICT DO NOTHING`, so it's safe to run multiple times
- GORM AutoMigrate handles the table schema, migrations only insert data
- The Chat-Only stream ID is hardcoded in `pkg/bot/manager.go`
