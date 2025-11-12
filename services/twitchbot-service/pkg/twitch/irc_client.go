package twitch

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

// IRCMessage represents a parsed IRC message
type IRCMessage struct {
	Raw     string
	Tags    map[string]string
	Prefix  string
	Command string
	Params  []string
}

// MessageHandler is called when a PRIVMSG is received
type MessageHandler func(channel, username, displayName, userID, message string, tags map[string]string)

// IRCClient represents a Twitch IRC client
type IRCClient struct {
	server         string
	port           string
	username       string
	oauth          string
	channel        string            // Primary channel
	channels       map[string]bool   // All joined channels
	channelsMu     sync.RWMutex      // Mutex for channels map
	conn           net.Conn
	reader         *textproto.Reader
	messageHandler MessageHandler
	connected      bool
	connectedSince time.Time
	tokenManager   *TokenManager // Optional token manager for auto-refresh
	debug          bool           // Enable debug logging
}

// NewIRCClient creates a new Twitch IRC client
func NewIRCClient(server, port, username, oauth, channel string) *IRCClient {
	return &IRCClient{
		server:   server,
		port:     port,
		username: username,
		oauth:    oauth,
		channel:  channel,
		channels: make(map[string]bool),
	}
}

// SetMessageHandler sets the handler for incoming messages
func (c *IRCClient) SetMessageHandler(handler MessageHandler) {
	c.messageHandler = handler
}

// SetTokenManager sets the token manager for automatic token refresh
func (c *IRCClient) SetTokenManager(tm *TokenManager) {
	c.tokenManager = tm

	// Set callback to reconnect when token is refreshed
	if tm != nil {
		tm.SetTokenRefreshCallback(func(newToken string) {
			log.Println("🔄 Token refreshed, reconnecting to IRC...")
			c.oauth = newToken
			// Trigger reconnect with new token
			if c.connected {
				ctx := context.Background()
				if err := c.reconnect(ctx); err != nil {
					log.Printf("❌ Failed to reconnect with new token: %v", err)
				}
			}
		})
	}
}

// EnableDebug enables debug logging for IRC messages
func (c *IRCClient) EnableDebug(enabled bool) {
	c.debug = enabled
	if enabled {
		log.Println("🐛 IRC Debug logging enabled")
	}
}

// Connect establishes connection to Twitch IRC
func (c *IRCClient) Connect(ctx context.Context) error {
	// Validate credentials
	if c.oauth == "" {
		return fmt.Errorf("oauth token is required")
	}
	if c.username == "" {
		return fmt.Errorf("username is required")
	}
	if c.channel == "" {
		return fmt.Errorf("channel is required")
	}

	// Normalize oauth token (remove "oauth:" prefix if present)
	// We'll add it back when sending to Twitch
	c.oauth = strings.TrimPrefix(c.oauth, "oauth:")

	address := fmt.Sprintf("%s:%s", c.server, c.port)
	log.Printf("Connecting to Twitch IRC at %s as %s...", address, c.username)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.reader = textproto.NewReader(bufio.NewReader(conn))

	// Request capabilities
	log.Println("   Requesting Twitch capabilities...")
	if err := c.send("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"); err != nil {
		return fmt.Errorf("failed to request capabilities: %w", err)
	}

	// Authenticate
	log.Println("   Authenticating...")
	if err := c.send(fmt.Sprintf("PASS oauth:%s", c.oauth)); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}
	if err := c.send(fmt.Sprintf("NICK %s", c.username)); err != nil {
		return fmt.Errorf("failed to send nick: %w", err)
	}

	// Wait for authentication to complete
	log.Println("   Waiting for authentication...")
	time.Sleep(2 * time.Second)

	// Join primary channel
	log.Printf("   Joining channel #%s...", c.channel)
	if err := c.send(fmt.Sprintf("JOIN #%s", c.channel)); err != nil {
		return fmt.Errorf("failed to join channel: %w", err)
	}

	// Track joined channel
	c.channelsMu.Lock()
	c.channels[c.channel] = true
	c.channelsMu.Unlock()

	c.connected = true
	c.connectedSince = time.Now()
	log.Printf("Successfully connected to Twitch IRC and joined #%s", c.channel)

	// Start message loop
	go c.messageLoop(ctx)

	return nil
}

// Disconnect closes the IRC connection
func (c *IRCClient) Disconnect() error {
	if c.conn != nil {
		c.connected = false
		c.send(fmt.Sprintf("PART #%s", c.channel))
		return c.conn.Close()
	}
	return nil
}

// SendMessage sends a message to the primary channel
func (c *IRCClient) SendMessage(message string) error {
	if !c.connected {
		return fmt.Errorf("not connected to IRC")
	}
	return c.send(fmt.Sprintf("PRIVMSG #%s :%s", c.channel, message))
}

// SendMessageToChannel sends a message to a specific channel
func (c *IRCClient) SendMessageToChannel(channel, message string) error {
	if !c.connected {
		return fmt.Errorf("not connected to IRC")
	}

	// Check if we're in this channel
	c.channelsMu.RLock()
	inChannel := c.channels[channel]
	c.channelsMu.RUnlock()

	if !inChannel {
		return fmt.Errorf("not joined to channel #%s", channel)
	}

	return c.send(fmt.Sprintf("PRIVMSG #%s :%s", channel, message))
}

// JoinChannel joins a new channel
func (c *IRCClient) JoinChannel(channel string) error {
	if !c.connected {
		return fmt.Errorf("not connected to IRC")
	}

	// Check if already joined
	c.channelsMu.RLock()
	alreadyJoined := c.channels[channel]
	c.channelsMu.RUnlock()

	if alreadyJoined {
		return fmt.Errorf("already joined to channel #%s", channel)
	}

	// Send JOIN command
	if err := c.send(fmt.Sprintf("JOIN #%s", channel)); err != nil {
		return fmt.Errorf("failed to join channel: %w", err)
	}

	// Track the channel
	c.channelsMu.Lock()
	c.channels[channel] = true
	c.channelsMu.Unlock()

	log.Printf("Joined channel #%s", channel)
	return nil
}

// LeaveChannel leaves a channel
func (c *IRCClient) LeaveChannel(channel string) error {
	if !c.connected {
		return fmt.Errorf("not connected to IRC")
	}

	// Don't allow leaving the primary channel
	if channel == c.channel {
		return fmt.Errorf("cannot leave primary channel #%s", channel)
	}

	// Check if we're in this channel
	c.channelsMu.RLock()
	inChannel := c.channels[channel]
	c.channelsMu.RUnlock()

	if !inChannel {
		return fmt.Errorf("not joined to channel #%s", channel)
	}

	// Send PART command
	if err := c.send(fmt.Sprintf("PART #%s", channel)); err != nil {
		return fmt.Errorf("failed to leave channel: %w", err)
	}

	// Remove from tracking
	c.channelsMu.Lock()
	delete(c.channels, channel)
	c.channelsMu.Unlock()

	log.Printf("Left channel #%s", channel)
	return nil
}

// GetJoinedChannels returns a list of all joined channels
func (c *IRCClient) GetJoinedChannels() []string {
	c.channelsMu.RLock()
	defer c.channelsMu.RUnlock()

	channels := make([]string, 0, len(c.channels))
	for ch := range c.channels {
		channels = append(channels, ch)
	}
	return channels
}

// IsConnected returns whether the client is connected
func (c *IRCClient) IsConnected() bool {
	return c.connected
}

// GetConnectedSince returns when the connection was established
func (c *IRCClient) GetConnectedSince() time.Time {
	return c.connectedSince
}

// GetPrimaryChannel returns the primary channel
func (c *IRCClient) GetPrimaryChannel() string {
	return c.channel
}

// GetUsername returns the bot username
func (c *IRCClient) GetUsername() string {
	return c.username
}

// send sends a raw IRC message
func (c *IRCClient) send(message string) error {
	if c.conn == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := fmt.Fprintf(c.conn, "%s\r\n", message)
	return err
}

// messageLoop continuously reads and processes IRC messages
func (c *IRCClient) messageLoop(ctx context.Context) {
	reconnectAttempts := 0
	maxReconnectAttempts := 5

	for c.connected {
		select {
		case <-ctx.Done():
			log.Println("IRC message loop stopped by context")
			return
		default:
			line, err := c.reader.ReadLine()
			if err != nil {
				if !c.connected {
					// Connection was intentionally closed
					return
				}

				log.Printf("Error reading IRC message: %v", err)

				// Connection lost, try to reconnect
				if reconnectAttempts < maxReconnectAttempts {
					reconnectAttempts++
					log.Printf("Connection lost, attempting reconnect %d/%d...", reconnectAttempts, maxReconnectAttempts)
					time.Sleep(time.Duration(reconnectAttempts) * 5 * time.Second)

					if err := c.reconnect(ctx); err != nil {
						log.Printf("Reconnect failed: %v", err)
						continue
					}

					log.Println("Successfully reconnected to IRC")
					reconnectAttempts = 0
				} else {
					log.Printf("Max reconnect attempts reached, giving up")
					c.connected = false
					return
				}
				continue
			}

			// Successfully read a line, reset reconnect counter
			reconnectAttempts = 0
			c.handleMessage(line)
		}
	}
}

// reconnect attempts to reconnect to IRC
func (c *IRCClient) reconnect(ctx context.Context) error {
	// Close old connection
	if c.conn != nil {
		c.conn.Close()
	}

	// Establish new connection
	address := fmt.Sprintf("%s:%s", c.server, c.port)
	log.Printf("   Reconnecting to %s...", address)
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	c.conn = conn
	c.reader = textproto.NewReader(bufio.NewReader(conn))

	// Re-authenticate (oauth token already normalized from initial Connect)
	log.Println("   Requesting capabilities...")
	if err := c.send("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"); err != nil {
		return fmt.Errorf("failed to request capabilities: %w", err)
	}

	log.Println("   Re-authenticating...")
	if err := c.send(fmt.Sprintf("PASS oauth:%s", c.oauth)); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}
	if err := c.send(fmt.Sprintf("NICK %s", c.username)); err != nil {
		return fmt.Errorf("failed to send nick: %w", err)
	}

	log.Println("   Waiting for re-authentication...")
	time.Sleep(2 * time.Second)

	log.Printf("   Re-joining channel #%s...", c.channel)
	if err := c.send(fmt.Sprintf("JOIN #%s", c.channel)); err != nil {
		return fmt.Errorf("failed to join channel: %w", err)
	}

	return nil
}

// handleMessage processes a single IRC message
func (c *IRCClient) handleMessage(raw string) {
	// Log raw IRC messages if debug enabled
	if c.debug {
		log.Printf("IRC< %s", raw)
	}

	msg := parseIRCMessage(raw)

	switch msg.Command {
	case "PING":
		// Respond to PING to keep connection alive
		if len(msg.Params) > 0 {
			c.send(fmt.Sprintf("PONG :%s", msg.Params[0]))
		}

	case "NOTICE":
		// Log important notices from Twitch
		if len(msg.Params) > 1 {
			notice := msg.Params[1]
			log.Printf("TWITCH NOTICE: %s", notice)

			// Check for authentication failures
			if strings.Contains(notice, "Login authentication failed") ||
			   strings.Contains(notice, "Improperly formatted auth") {
				log.Println("❌ IRC Authentication failed!")
				log.Println("   Check that:")
				log.Println("   1. TWITCH_ACCESS_TOKEN is valid")
				log.Println("   2. TWITCH_BOT_USERNAME matches the token's user")
				log.Println("   3. Token has chat:read and chat:edit scopes")
			}
		}

	case "PRIVMSG":
		// Handle chat message
		if len(msg.Params) >= 2 && c.messageHandler != nil {
			channel := msg.Params[0]
			message := msg.Params[1]

			// Extract user information from prefix and tags
			username := extractUsername(msg.Prefix)
			displayName := msg.Tags["display-name"]
			if displayName == "" {
				displayName = username
			}
			userID := msg.Tags["user-id"]

			c.messageHandler(channel, username, displayName, userID, message, msg.Tags)
		}

	case "001": // Welcome message
		log.Println("Successfully authenticated with Twitch IRC")

	case "JOIN":
		log.Printf("Joined channel: %s", msg.Params[0])

	case "PART":
		log.Printf("Left channel: %s", msg.Params[0])
	}
}

// parseIRCMessage parses a raw IRC message into structured data
func parseIRCMessage(raw string) IRCMessage {
	msg := IRCMessage{
		Raw:  raw,
		Tags: make(map[string]string),
	}

	// Parse tags (if present)
	if strings.HasPrefix(raw, "@") {
		parts := strings.SplitN(raw[1:], " ", 2)
		tagString := parts[0]
		raw = parts[1]

		for _, tag := range strings.Split(tagString, ";") {
			kv := strings.SplitN(tag, "=", 2)
			if len(kv) == 2 {
				msg.Tags[kv[0]] = kv[1]
			}
		}
	}

	// Parse prefix (if present)
	if strings.HasPrefix(raw, ":") {
		parts := strings.SplitN(raw[1:], " ", 2)
		msg.Prefix = parts[0]
		raw = parts[1]
	}

	// Parse command and params
	parts := strings.SplitN(raw, " :", 2)
	commandAndParams := strings.Fields(parts[0])

	if len(commandAndParams) > 0 {
		msg.Command = commandAndParams[0]
		if len(commandAndParams) > 1 {
			msg.Params = commandAndParams[1:]
		}
	}

	// Add trailing parameter if present
	if len(parts) > 1 {
		msg.Params = append(msg.Params, parts[1])
	}

	return msg
}

// extractUsername extracts username from IRC prefix
func extractUsername(prefix string) string {
	if strings.Contains(prefix, "!") {
		return strings.Split(prefix, "!")[0]
	}
	return prefix
}
