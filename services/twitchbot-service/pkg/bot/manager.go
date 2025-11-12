package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"toxictoast/services/twitchbot-service/internal/usecase"
	"toxictoast/services/twitchbot-service/pkg/config"
	"toxictoast/services/twitchbot-service/pkg/events"
	"toxictoast/services/twitchbot-service/pkg/twitch"
)

// Manager coordinates the Twitch bot components
type Manager struct {
	cfg              *config.Config
	ircClient        *twitch.IRCClient
	helixClient      *twitch.HelixClient
	tokenManager     *twitch.TokenManager
	streamUC         usecase.StreamUseCase
	messageUC        usecase.MessageUseCase
	viewerUC         usecase.ViewerUseCase
	clipUC           usecase.ClipUseCase
	commandUC        usecase.CommandUseCase
	channelViewerUC  usecase.ChannelViewerUseCase
	eventPublisher   *events.EventPublisher
	activeStreamIDs  map[string]string // channel -> streamID
	streamIDsMu      sync.RWMutex      // Mutex for activeStreamIDs
	broadcasterCache map[string]string // channel -> broadcasterID cache
	cacheMu          sync.RWMutex      // Mutex for broadcasterCache
	stopChan         chan struct{}
}

// NewManager creates a new bot manager
func NewManager(
	cfg *config.Config,
	streamUC usecase.StreamUseCase,
	messageUC usecase.MessageUseCase,
	viewerUC usecase.ViewerUseCase,
	clipUC usecase.ClipUseCase,
	commandUC usecase.CommandUseCase,
	channelViewerUC usecase.ChannelViewerUseCase,
	eventPublisher *events.EventPublisher,
) *Manager {
	return &Manager{
		cfg:              cfg,
		streamUC:         streamUC,
		messageUC:        messageUC,
		viewerUC:         viewerUC,
		clipUC:           clipUC,
		commandUC:        commandUC,
		channelViewerUC:  channelViewerUC,
		eventPublisher:   eventPublisher,
		activeStreamIDs:  make(map[string]string),
		broadcasterCache: make(map[string]string),
		stopChan:         make(chan struct{}),
	}
}

// Start initializes and starts the bot
func (m *Manager) Start(ctx context.Context) error {
	// Check if Twitch is configured
	if m.cfg.Twitch.Channel == "" {
		log.Println("⚠️  TWITCH_CHANNEL not configured - bot will not start")
		log.Println("ℹ️  Service is running in API-only mode (gRPC endpoints available)")
		log.Println("ℹ️  To enable bot: Set TWITCH_CHANNEL, TWITCH_ACCESS_TOKEN, TWITCH_BOT_USERNAME in .env")
		return nil
	}

	if m.cfg.Twitch.AccessToken == "" {
		log.Println("⚠️  TWITCH_ACCESS_TOKEN not configured - bot will not start")
		log.Println("ℹ️  Service is running in API-only mode (gRPC endpoints available)")
		return nil
	}

	if m.cfg.Twitch.BotUsername == "" {
		log.Println("⚠️  TWITCH_BOT_USERNAME not configured - bot will not start")
		log.Println("ℹ️  Service is running in API-only mode (gRPC endpoints available)")
		return nil
	}

	log.Println("🤖 Starting Twitch bot manager...")
	log.Printf("   Channel: #%s", m.cfg.Twitch.Channel)
	log.Printf("   Bot Username: %s", m.cfg.Twitch.BotUsername)

	// Initialize Token Manager for automatic token refresh
	if m.cfg.Twitch.ClientID != "" && m.cfg.Twitch.ClientSecret != "" {
		log.Println("   Initializing token manager with auto-refresh...")
		m.tokenManager = twitch.NewTokenManager(
			m.cfg.Twitch.ClientID,
			m.cfg.Twitch.ClientSecret,
			m.cfg.Twitch.AccessToken,
			"", // No refresh token initially
		)

		// Validate token and get expiration info
		tokenUser, err := m.tokenManager.ValidateToken()
		if err != nil {
			log.Printf("⚠️  Warning: Token validation failed: %v", err)
			log.Println("   Service will continue, but token refresh may not work")
		} else {
			// Check if token user matches bot username
			if tokenUser != "" && tokenUser != m.cfg.Twitch.BotUsername {
				log.Printf("⚠️  WARNING: Token user (%s) does not match bot username (%s)!", tokenUser, m.cfg.Twitch.BotUsername)
				log.Println("   This will cause authentication failures!")
				log.Printf("   Generate a new token for the account: %s", m.cfg.Twitch.BotUsername)
			}
		}
	} else {
		log.Println("ℹ️  Client ID/Secret not configured - token auto-refresh disabled")
	}

	// Initialize Helix client
	m.helixClient = twitch.NewHelixClient(
		m.cfg.Twitch.ClientID,
		m.cfg.Twitch.ClientSecret,
		m.cfg.Twitch.AccessToken,
	)

	// Set token manager on Helix client
	if m.tokenManager != nil {
		m.helixClient.SetTokenManager(m.tokenManager)
	}

	// Initialize IRC client
	m.ircClient = twitch.NewIRCClient(
		m.cfg.Twitch.IRCServer,
		m.cfg.Twitch.IRCPort,
		m.cfg.Twitch.BotUsername,
		m.cfg.Twitch.AccessToken,
		m.cfg.Twitch.Channel,
	)

	// Set token manager on IRC client
	if m.tokenManager != nil {
		m.ircClient.SetTokenManager(m.tokenManager)
	}

	// Enable debug logging if configured
	if m.cfg.Twitch.IRCDebug {
		m.ircClient.EnableDebug(true)
	}

	// Set up message handler
	m.ircClient.SetMessageHandler(m.handleChatMessage)

	// Connect to IRC
	log.Println("   Connecting to Twitch IRC...")
	if err := m.ircClient.Connect(ctx); err != nil {
		log.Printf("❌ Failed to connect to Twitch IRC: %v", err)
		log.Println("ℹ️  Common issues:")
		log.Println("   - Invalid TWITCH_ACCESS_TOKEN (get from https://twitchtokengenerator.com/)")
		log.Println("   - Wrong TWITCH_BOT_USERNAME (must match token)")
		log.Println("   - Token expired (generate new token)")
		log.Println("ℹ️  Service continues in API-only mode")
		return err
	}

	// Start stream watcher
	go m.streamWatcher(ctx)

	log.Println("✅ Twitch bot manager started successfully")
	log.Println("   - IRC connected")
	log.Println("   - Stream watcher running")
	log.Println("   - Ready to log messages")
	return nil
}

// Stop gracefully stops the bot
func (m *Manager) Stop() error {
	log.Println("Stopping Twitch bot manager...")
	close(m.stopChan)

	if m.ircClient != nil {
		if err := m.ircClient.Disconnect(); err != nil {
			log.Printf("Error disconnecting IRC: %v", err)
		}
	}

	log.Println("Twitch bot manager stopped")
	return nil
}

// JoinChannel joins a new channel and checks if it's live
func (m *Manager) JoinChannel(channel string) error {
	if m.ircClient == nil {
		return fmt.Errorf("bot is not running")
	}

	// Join the IRC channel first
	if err := m.ircClient.JoinChannel(channel); err != nil {
		return err
	}

	// Check if this channel is currently streaming
	if m.helixClient != nil {
		go m.checkChannelStreamStatus(context.Background(), channel)
	}

	// Fetch all current viewers in the channel
	if m.helixClient != nil && m.channelViewerUC != nil {
		go m.fetchChannelViewers(context.Background(), channel)
	}

	return nil
}

// LeaveChannel leaves a channel
func (m *Manager) LeaveChannel(channel string) error {
	if m.ircClient == nil {
		return fmt.Errorf("bot is not running")
	}

	// Leave the IRC channel
	if err := m.ircClient.LeaveChannel(channel); err != nil {
		return err
	}

	// Remove from active streams map
	m.streamIDsMu.Lock()
	delete(m.activeStreamIDs, channel)
	m.streamIDsMu.Unlock()

	return nil
}

// GetJoinedChannels returns all joined channels
func (m *Manager) GetJoinedChannels() []string {
	if m.ircClient == nil {
		return []string{}
	}
	return m.ircClient.GetJoinedChannels()
}

// GetPrimaryChannel returns the primary channel
func (m *Manager) GetPrimaryChannel() string {
	if m.ircClient == nil {
		return ""
	}
	return m.ircClient.GetPrimaryChannel()
}

// BotStatus represents the bot's current status
type BotStatus struct {
	Connected       bool
	Authenticated   bool
	ChannelsJoined  int
	ActiveChannels  []string
	BotUsername     string
	ConnectedSince  time.Time
}

// GetStatus returns the bot's current status
func (m *Manager) GetStatus() BotStatus {
	if m.ircClient == nil {
		return BotStatus{
			Connected:     false,
			Authenticated: false,
		}
	}

	return BotStatus{
		Connected:       m.ircClient.IsConnected(),
		Authenticated:   m.ircClient.IsConnected(), // If connected, we're authenticated
		ChannelsJoined:  len(m.ircClient.GetJoinedChannels()),
		ActiveChannels:  m.ircClient.GetJoinedChannels(),
		BotUsername:     m.ircClient.GetUsername(),
		ConnectedSince:  m.ircClient.GetConnectedSince(),
	}
}

// SendMessage sends a message to a channel (or primary if channel is empty)
func (m *Manager) SendMessage(channel, message string) error {
	if m.ircClient == nil {
		return fmt.Errorf("bot is not running")
	}

	if channel == "" {
		// Send to primary channel
		return m.ircClient.SendMessage(message)
	}

	// Send to specific channel
	return m.ircClient.SendMessageToChannel(channel, message)
}

// handleChatMessage processes incoming chat messages
func (m *Manager) handleChatMessage(channel, username, displayName, userID, message string, tags map[string]string) {
	ctx := context.Background()

	// Fixed UUID for Chat-Only stream (matches migration 002)
	const chatOnlyStreamID = "00000000-0000-0000-0000-000000000001"

	// Remove # prefix from channel if present
	channel = strings.TrimPrefix(channel, "#")

	// Get or find active stream for this channel
	m.streamIDsMu.RLock()
	streamID := m.activeStreamIDs[channel]
	m.streamIDsMu.RUnlock()

	if streamID == "" {
		// No stream tracked for this channel yet
		// Check if this channel is currently live
		m.streamIDsMu.Lock()
		if m.helixClient != nil {
			streamData, err := m.helixClient.GetStream(channel)
			if err == nil && streamData != nil && streamData.Type == "live" {
				// Channel is live! Create a stream for it
				log.Printf("📺 Channel %s is live, creating stream: %s", channel, streamData.Title)
				stream, err := m.streamUC.CreateStream(ctx, streamData.Title, streamData.GameName, streamData.GameID)
				if err == nil {
					streamID = stream.ID
					m.activeStreamIDs[channel] = streamID
					log.Printf("✅ Stream created for %s: %s", channel, streamID)
				}
			}
		}

		// If still no stream (channel offline or API failed), use Chat-Only
		if streamID == "" {
			streamID = chatOnlyStreamID
			m.activeStreamIDs[channel] = streamID
			log.Printf("📝 Channel %s offline, using Chat-Only stream", channel)
		}
		m.streamIDsMu.Unlock()
	}

	// Check if message is a command
	if len(message) > 0 && message[0] == '!' {
		m.handleCommand(username, userID, message, tags)
		return
	}

	// Parse tags for user info
	isModerator := tags["mod"] == "1"
	isSubscriber := tags["subscriber"] == "1"
	isVIP := tags["vip"] == "1"
	isBroadcaster := tags["badges"] != "" && tags["badges"][:10] == "broadcaster"

	// Track viewer in this channel (updates last_seen and badges)
	if m.channelViewerUC != nil && userID != "" {
		_, err := m.channelViewerUC.AddViewer(
			ctx,
			channel,
			userID,
			username,
			displayName,
			isModerator,
			isVIP,
		)
		if err != nil {
			log.Printf("⚠️  Failed to track viewer %s in channel %s: %v", username, channel, err)
		}
	}

	// Create message in database
	msg, err := m.messageUC.CreateMessage(
		ctx,
		streamID,
		userID,
		username,
		displayName,
		message,
		isModerator,
		isSubscriber,
		isVIP,
		isBroadcaster,
	)

	if err != nil {
		log.Printf("Error creating message: %v", err)
		return
	}

	// Publish event
	if m.eventPublisher != nil {
		m.eventPublisher.PublishMessageEvent(events.MessageEvent{
			BaseEvent: events.BaseEvent{
				Type: events.MessageReceived,
			},
			MessageID:     msg.ID,
			StreamID:      msg.StreamID,
			UserID:        msg.UserID,
			Username:      msg.Username,
			DisplayName:   msg.DisplayName,
			Message:       msg.Message,
			IsModerator:   msg.IsModerator,
			IsSubscriber:  msg.IsSubscriber,
			IsVIP:         msg.IsVIP,
			IsBroadcaster: msg.IsBroadcaster,
		})
	}
}

// handleCommand processes chat commands
func (m *Manager) handleCommand(username, userID, message string, tags map[string]string) {
	commandName := message
	if idx := len(message); idx > 0 {
		// Extract command name (first word)
		for i, r := range message {
			if r == ' ' {
				commandName = message[:i]
				break
			}
		}
	}

	isModerator := tags["mod"] == "1"
	isSubscriber := tags["subscriber"] == "1"

	ctx := context.Background()
	success, response, err, _ := m.commandUC.ExecuteCommand(
		ctx,
		commandName,
		userID,
		username,
		isModerator,
		isSubscriber,
	)

	if err != nil {
		log.Printf("Command execution error: %v", err)
		return
	}

	if success && response != "" {
		// Send response to chat
		if m.ircClient != nil {
			if err := m.ircClient.SendMessage(response); err != nil {
				log.Printf("Error sending command response: %v", err)
			}
		}
	}
}

// streamWatcher polls Helix API for stream status
func (m *Manager) streamWatcher(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("Stream watcher started")

	// Initial check for primary channel
	if m.cfg.Twitch.Channel != "" {
		m.checkChannelStreamStatus(ctx, m.cfg.Twitch.Channel)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stream watcher stopped by context")
			return
		case <-m.stopChan:
			log.Println("Stream watcher stopped")
			return
		case <-ticker.C:
			// Check all joined channels
			m.checkAllChannelStreams(ctx)
		}
	}
}

// checkAllChannelStreams checks stream status for all joined channels
func (m *Manager) checkAllChannelStreams(ctx context.Context) {
	if m.ircClient == nil {
		return
	}

	// Get all joined channels
	channels := m.ircClient.GetJoinedChannels()
	for _, channel := range channels {
		m.checkChannelStreamStatus(ctx, channel)
	}
}

// checkChannelStreamStatus checks if a specific channel is streaming
func (m *Manager) checkChannelStreamStatus(ctx context.Context, channel string) {
	const chatOnlyStreamID = "00000000-0000-0000-0000-000000000001"

	if m.helixClient == nil {
		return
	}

	// Remove # prefix if present
	channel = strings.TrimPrefix(channel, "#")

	// Get stream from Helix API
	streamData, err := m.helixClient.GetStream(channel)
	if err != nil {
		log.Printf("Error fetching stream status for %s: %v", channel, err)
		return
	}

	// Get current stream ID for this channel
	m.streamIDsMu.RLock()
	currentStreamID := m.activeStreamIDs[channel]
	m.streamIDsMu.RUnlock()

	// Check if stream is live
	if streamData != nil && streamData.Type == "live" {
		// Channel is live
		if currentStreamID == "" || currentStreamID == chatOnlyStreamID {
			// New stream started (switch from Chat-Only to real stream)
			m.handleStreamStart(ctx, channel, streamData)
		} else {
			// Update existing stream
			m.handleStreamUpdate(ctx, channel, currentStreamID, streamData)
		}
	} else {
		// Stream is offline
		if currentStreamID != "" && currentStreamID != chatOnlyStreamID {
			// Real stream ended (switch back to Chat-Only)
			m.handleStreamEnd(ctx, channel, currentStreamID)
		}
	}
}

// handleStreamStart creates a new stream record for a channel
func (m *Manager) handleStreamStart(ctx context.Context, channel string, streamData *twitch.StreamData) {
	log.Printf("Stream started on %s: %s", channel, streamData.Title)

	stream, err := m.streamUC.CreateStream(ctx, streamData.Title, streamData.GameName, streamData.GameID)
	if err != nil {
		log.Printf("Error creating stream for %s: %v", channel, err)
		return
	}

	// Update stream ID for this channel
	m.streamIDsMu.Lock()
	m.activeStreamIDs[channel] = stream.ID
	m.streamIDsMu.Unlock()

	log.Printf("✅ Stream created for %s: %s (ID: %s)", channel, stream.Title, stream.ID)

	// Publish event
	if m.eventPublisher != nil {
		m.eventPublisher.PublishStreamEvent(events.StreamEvent{
			BaseEvent: events.BaseEvent{
				Type: events.StreamStarted,
			},
			StreamID:    stream.ID,
			Title:       stream.Title,
			GameName:    stream.GameName,
			GameID:      stream.GameID,
			ViewerCount: streamData.ViewerCount,
			IsActive:    true,
			StartedAt:   stream.StartedAt.Format(time.RFC3339),
		})
	}
}

// handleStreamUpdate updates stream viewer count for a channel
func (m *Manager) handleStreamUpdate(ctx context.Context, channel string, streamID string, streamData *twitch.StreamData) {
	// Update peak viewers if current is higher
	stream, err := m.streamUC.GetStreamByID(ctx, streamID)
	if err != nil {
		return
	}

	peakViewers := stream.PeakViewers
	if streamData.ViewerCount > peakViewers {
		peakViewers = streamData.ViewerCount
	}

	// Simple average calculation (could be improved)
	averageViewers := (stream.AverageViewers + streamData.ViewerCount) / 2

	_, err = m.streamUC.UpdateStream(ctx, streamID, nil, nil, nil, &peakViewers, &averageViewers)
	if err != nil {
		log.Printf("Error updating stream for %s: %v", channel, err)
	}
}

// handleStreamEnd marks the stream as ended for a channel
func (m *Manager) handleStreamEnd(ctx context.Context, channel string, streamID string) {
	const chatOnlyStreamID = "00000000-0000-0000-0000-000000000001"

	log.Printf("Stream ended on %s", channel)

	stream, err := m.streamUC.EndStream(ctx, streamID)
	if err != nil {
		log.Printf("Error ending stream for %s: %v", channel, err)
		return
	}

	// Publish event
	if m.eventPublisher != nil {
		endedAt := ""
		if stream.EndedAt != nil {
			endedAt = stream.EndedAt.Format(time.RFC3339)
		}

		m.eventPublisher.PublishStreamEvent(events.StreamEvent{
			BaseEvent: events.BaseEvent{
				Type: events.StreamEnded,
			},
			StreamID:       stream.ID,
			Title:          stream.Title,
			GameName:       stream.GameName,
			GameID:         stream.GameID,
			IsActive:       false,
			StartedAt:      stream.StartedAt.Format(time.RFC3339),
			EndedAt:        endedAt,
			PeakViewers:    stream.PeakViewers,
			AverageViewers: stream.AverageViewers,
			TotalMessages:  stream.TotalMessages,
		})
	}

	// Switch back to Chat-Only stream for this channel
	m.streamIDsMu.Lock()
	m.activeStreamIDs[channel] = chatOnlyStreamID
	m.streamIDsMu.Unlock()

	log.Printf("📝 Channel %s switched to Chat-Only stream", channel)
}

// fetchChannelViewers fetches all current viewers in a channel and adds them to the database
func (m *Manager) fetchChannelViewers(ctx context.Context, channel string) {
	if m.helixClient == nil || m.channelViewerUC == nil {
		return
	}

	// Remove # prefix if present
	channel = strings.TrimPrefix(channel, "#")

	log.Printf("🔍 Fetching viewers for channel: %s", channel)

	// Get broadcaster ID (with cache)
	broadcasterID, err := m.getBroadcasterID(ctx, channel)
	if err != nil {
		log.Printf("⚠️  Failed to get broadcaster ID for %s: %v", channel, err)
		return
	}

	// Get bot's user ID (moderator ID required for chatters endpoint)
	botUser, err := m.helixClient.GetUser(m.cfg.Twitch.BotUsername)
	if err != nil {
		log.Printf("⚠️  Failed to get bot user ID: %v", err)
		return
	}

	// Fetch chatters from Helix API
	chatters, err := m.helixClient.GetChatters(broadcasterID, botUser.ID)
	if err != nil {
		log.Printf("⚠️  Failed to fetch chatters for %s: %v", channel, err)
		log.Printf("   Note: Bot requires 'moderator:read:chatters' scope and moderator privileges")
		return
	}

	log.Printf("✅ Found %d viewers in %s, adding to database...", len(chatters), channel)

	// Add each chatter to the channel viewer table
	addedCount := 0
	for _, chatter := range chatters {
		// Note: We don't have moderator/VIP status from chatters endpoint
		// These will be updated as users send messages with their badges
		_, err := m.channelViewerUC.AddViewer(
			ctx,
			channel,
			chatter.UserID,
			chatter.UserLogin,
			chatter.UserName,
			false, // isMod - will be updated from message tags
			false, // isVIP - will be updated from message tags
		)
		if err != nil {
			log.Printf("⚠️  Failed to add viewer %s to %s: %v", chatter.UserLogin, channel, err)
		} else {
			addedCount++
		}
	}

	log.Printf("✅ Added %d/%d viewers to channel %s", addedCount, len(chatters), channel)
}

// getBroadcasterID gets the broadcaster ID for a channel (with caching)
func (m *Manager) getBroadcasterID(ctx context.Context, channel string) (string, error) {
	// Check cache first
	m.cacheMu.RLock()
	broadcasterID, ok := m.broadcasterCache[channel]
	m.cacheMu.RUnlock()

	if ok && broadcasterID != "" {
		return broadcasterID, nil
	}

	// Not in cache, fetch from API
	user, err := m.helixClient.GetUser(channel)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	// Cache the result
	m.cacheMu.Lock()
	m.broadcasterCache[channel] = user.ID
	m.cacheMu.Unlock()

	return user.ID, nil
}
