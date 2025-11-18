-- Create a permanent "Chat-Only" stream for logging messages when not live
-- This stream uses a fixed UUID so we can reference it easily

-- Insert the Chat-Only stream with a fixed UUID
INSERT INTO streams (
    id,
    title,
    game_name,
    game_id,
    started_at,
    ended_at,
    peak_viewers,
    average_viewers,
    total_messages,
    is_active,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    '00000000-0000-0000-0000-000000000001',  -- Fixed UUID for Chat-Only stream
    'Chat-Only Messages',                      -- Title
    'Just Chatting',                           -- Game name
    '509658',                                  -- Just Chatting game ID
    NOW(),                                     -- Started at (now)
    NULL,                                      -- Never ended
    0,                                         -- Peak viewers
    0,                                         -- Average viewers
    0,                                         -- Total messages (will increment)
    true,                                      -- Always active
    NOW(),                                     -- Created at
    NOW(),                                     -- Updated at
    NULL                                       -- Not deleted
) ON CONFLICT (id) DO NOTHING;  -- Don't insert if already exists

-- Create an index for faster lookups of the chat-only stream
CREATE INDEX IF NOT EXISTS idx_streams_chat_only
ON streams(id)
WHERE id = '00000000-0000-0000-0000-000000000001';
