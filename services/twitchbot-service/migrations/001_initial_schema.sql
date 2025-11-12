-- Initial schema for twitchbot-service
-- This file is for documentation purposes
-- Migrations are handled automatically by GORM AutoMigrate

-- Streams table
CREATE TABLE IF NOT EXISTS streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    game_name VARCHAR(255),
    game_id VARCHAR(100),
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    peak_viewers INTEGER DEFAULT 0,
    average_viewers INTEGER DEFAULT 0,
    total_messages INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_streams_deleted_at ON streams(deleted_at);
CREATE INDEX idx_streams_is_active ON streams(is_active);
CREATE INDEX idx_streams_started_at ON streams(started_at);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    is_moderator BOOLEAN DEFAULT false,
    is_subscriber BOOLEAN DEFAULT false,
    is_vip BOOLEAN DEFAULT false,
    is_broadcaster BOOLEAN DEFAULT false,
    sent_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_stream_id ON messages(stream_id);
CREATE INDEX idx_messages_user_id ON messages(user_id);
CREATE INDEX idx_messages_sent_at ON messages(sent_at);
CREATE INDEX idx_messages_deleted_at ON messages(deleted_at);

-- Viewers table
CREATE TABLE IF NOT EXISTS viewers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    twitch_id VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    total_messages INTEGER DEFAULT 0,
    total_streams_watched INTEGER DEFAULT 0,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_viewers_twitch_id ON viewers(twitch_id);
CREATE INDEX idx_viewers_deleted_at ON viewers(deleted_at);
CREATE INDEX idx_viewers_last_seen ON viewers(last_seen);

-- Clips table
CREATE TABLE IF NOT EXISTS clips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL,
    twitch_clip_id VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    embed_url TEXT,
    thumbnail_url TEXT,
    creator_name VARCHAR(100) NOT NULL,
    creator_id VARCHAR(100) NOT NULL,
    view_count INTEGER DEFAULT 0,
    duration_seconds INTEGER NOT NULL,
    created_at_twitch TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE
);

CREATE INDEX idx_clips_stream_id ON clips(stream_id);
CREATE INDEX idx_clips_twitch_clip_id ON clips(twitch_clip_id);
CREATE INDEX idx_clips_creator_id ON clips(creator_id);
CREATE INDEX idx_clips_deleted_at ON clips(deleted_at);

-- Commands table
CREATE TABLE IF NOT EXISTS commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    response TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    moderator_only BOOLEAN DEFAULT false,
    subscriber_only BOOLEAN DEFAULT false,
    cooldown_seconds INTEGER DEFAULT 0,
    usage_count INTEGER DEFAULT 0,
    last_used TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_commands_name ON commands(name);
CREATE INDEX idx_commands_is_active ON commands(is_active);
CREATE INDEX idx_commands_deleted_at ON commands(deleted_at);
