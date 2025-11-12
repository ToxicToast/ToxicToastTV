# Twitch Bot Setup Guide

This guide explains how to configure the Twitch bot to connect to your channel.

## Quick Start (Without Bot)

The service works without Twitch credentials! It runs in **API-only mode**:
- ✅ All gRPC endpoints work
- ✅ Manual data entry via API
- ❌ No automatic chat logging
- ❌ No stream tracking

Just leave the Twitch variables empty in `.env` and start the service.

---

## Full Setup (With Bot)

To enable automatic chat logging and stream tracking, you need Twitch credentials.

### Step 1: Get Your Access Token

**Option A: Quick & Easy (Recommended)**

1. Go to https://twitchtokengenerator.com/
2. Select **Bot Chat Token**
3. Click **Generate Token**
4. Login with your **bot account** (not your streamer account!)
5. Authorize the app
6. Copy the **Access Token** (starts with `oauth:`)

**Option B: Manual (Advanced)**

1. Create app at https://dev.twitch.tv/console/apps
2. Get Client ID & Secret
3. Generate OAuth token manually
4. Needs scopes: `chat:read`, `chat:edit`, `channel:read:subscriptions`

### Step 2: Configure .env

Edit `services/twitchbot-service/.env`:

```env
# Your Twitch channel name (without #)
TWITCH_CHANNEL=your_channel_name

# Bot username (the account that will appear in chat)
TWITCH_BOT_USERNAME=your_bot_username

# Access token from twitchtokengenerator.com
# Note: "oauth:" prefix is optional - we'll handle it automatically
TWITCH_ACCESS_TOKEN=your_token_here

# Optional: For Helix API (clips, stream info)
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret
```

### Step 3: Start the Service

```bash
cd services/twitchbot-service
go run cmd/server/main.go
```

You should see:
```
🤖 Starting Twitch bot manager...
   Channel: #your_channel
   Bot Username: your_bot
   Connecting to Twitch IRC...
✅ Twitch bot manager started successfully
   - IRC connected
   - Stream watcher running
   - Ready to log messages
```

---

## Troubleshooting

### Error: "oauth token is required"
➡️ You forgot to set `TWITCH_ACCESS_TOKEN` in `.env`

### Error: "EOF" or "connection reset"
This usually means Twitch rejected your authentication. Common causes:

1. **Token doesn't match bot username**
   ```
   ⚠️  WARNING: Token user (streamer123) does not match bot username (bot456)!
   ```
   ➡️ **Solution**: Generate a token for the EXACT account specified in `TWITCH_BOT_USERNAME`

2. **Missing scopes** (chat:read, chat:edit)
   ```
   ⚠️  WARNING: Token is missing required scopes: [chat:edit]
   ```
   ➡️ **Solution**: Generate new token at twitchtokengenerator.com with "Bot Chat Token" selected

3. **Expired token**
   ```
   ⚠️  Access token is invalid or expired
   ```
   ➡️ **Solution**: Generate new token (or enable auto-refresh with Client ID/Secret)

### Enable Debug Logging

If you're still getting EOF errors, enable IRC debug logging:

```env
TWITCH_IRC_DEBUG=true
```

This will show ALL IRC messages, including authentication failures:
```
TWITCH NOTICE: Login authentication failed
❌ IRC Authentication failed!
   Check that:
   1. TWITCH_ACCESS_TOKEN is valid
   2. TWITCH_BOT_USERNAME matches the token's user
   3. Token has chat:read and chat:edit scopes
```

### Bot connects but doesn't log messages
➡️ Check:
- Is your stream live? Bot only logs when stream is active
- Database connection working? Check PostgreSQL logs
- Stream created? Bot auto-creates stream when going live

### "Service is running in API-only mode"
➡️ This is normal! It means:
- Bot is disabled (missing credentials)
- gRPC API still works
- You can manually create streams/messages via API

---

## Testing Without Going Live

You can test the bot without streaming:

### 1. API-Only Testing
```bash
# Create a test stream via gRPC
grpcurl -plaintext -d '{
  "title": "Test Stream",
  "game_name": "Just Chatting"
}' localhost:9093 twitchbot.StreamService/CreateStream

# Create a test message
grpcurl -plaintext -d '{
  "stream_id": "stream-id-from-above",
  "user_id": "12345",
  "username": "testuser",
  "display_name": "TestUser",
  "message": "Hello from API!"
}' localhost:9093 twitchbot.MessageService/CreateMessage
```

### 2. IRC Bot Testing (No Stream Required)
The bot connects to chat even when you're offline!
- Bot joins your channel
- Bot sees all chat messages
- Bot logs messages when stream is live
- When you go live, bot auto-creates stream record

---

## What the Bot Does Automatically

### When Stream Goes Live:
1. ✅ Creates `Stream` record in database
2. ✅ Publishes `stream.started` Kafka event
3. ✅ Starts logging all chat messages
4. ✅ Updates viewer count every 30 seconds

### During Stream:
1. ✅ Logs every chat message to database
2. ✅ Publishes `message.received` events
3. ✅ Tracks viewer engagement
4. ✅ Updates peak/average viewers
5. ✅ Executes chat commands (`!hello`, etc.)

### When Stream Ends:
1. ✅ Marks stream as ended
2. ✅ Publishes `stream.ended` event
3. ✅ Calculates final statistics
4. ✅ Stops logging messages (bot stays connected)

---

## Bot Architecture

```
┌─────────────────────────────────────────┐
│          Twitch Bot Manager             │
├─────────────────────────────────────────┤
│                                         │
│  ┌────────────┐      ┌──────────────┐  │
│  │ IRC Client │      │ Helix Client │  │
│  │ (Chat)     │      │ (API)        │  │
│  └────────────┘      └──────────────┘  │
│        │                    │           │
│        v                    v           │
│  ┌──────────────────────────────────┐  │
│  │      Message Handler             │  │
│  │  - Logs messages to DB           │  │
│  │  - Executes commands             │  │
│  │  - Publishes Kafka events        │  │
│  └──────────────────────────────────┘  │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │      Stream Watcher              │  │
│  │  - Polls every 30 seconds        │  │
│  │  - Detects stream start/end      │  │
│  │  - Updates viewer counts         │  │
│  └──────────────────────────────────┘  │
│                                         │
└─────────────────────────────────────────┘
```

---

## Security Notes

⚠️ **NEVER commit your `.env` file to git!**

- `.env` is in `.gitignore`
- Use `.env.example` as template
- Rotate tokens regularly
- Use separate bot account (not streamer account)

---

## Automatic Token Refresh

The service includes **automatic token refresh** functionality:

### How It Works

If you provide `TWITCH_CLIENT_ID` and `TWITCH_CLIENT_SECRET`:
- ✅ Token is automatically validated on startup
- ✅ Token expiration time is tracked
- ✅ Token is refreshed automatically before it expires
- ✅ IRC client reconnects with new token
- ✅ API requests automatically use fresh token

### Setup for Auto-Refresh

```env
# Required for auto-refresh
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret

# Your access token (will be refreshed automatically)
TWITCH_ACCESS_TOKEN=your_token_here
```

### Without Auto-Refresh

If you don't provide Client ID/Secret:
- ⚠️ Token will eventually expire
- ⚠️ You'll need to manually update `.env` with a new token
- ℹ️ Service will show "Client ID/Secret not configured" message

### Token Validation

On startup, you'll see:
```
✅ Token validated successfully (expires in 14400 seconds)
   User: your_bot_username (ID: 123456789)
   Scopes: [chat:read chat:edit]
```

If token is about to expire or invalid:
```
🔄 Refreshing access token...
✅ Token refreshed using client credentials
✅ New token expires in 3600 seconds
🔄 Token refreshed, reconnecting to IRC...
```

---

## Support

Common issues are usually:
1. **Token expired** - Will auto-refresh if Client ID/Secret configured
2. **Wrong username** - Must match token
3. **Token format** - `oauth:` prefix is optional (handled automatically)
4. **Auto-refresh not working** - Check Client ID/Secret are set correctly

For more help, check service logs for detailed error messages.
