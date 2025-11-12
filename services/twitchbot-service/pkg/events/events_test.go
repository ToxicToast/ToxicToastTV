package events

import (
	"testing"
	"time"
)

func TestEventTypes(t *testing.T) {
	tests := []struct {
		name     string
		event    EventType
		expected string
	}{
		{"Stream Started", StreamStarted, "twitchbot.stream.started"},
		{"Stream Ended", StreamEnded, "twitchbot.stream.ended"},
		{"Stream Updated", StreamUpdated, "twitchbot.stream.updated"},
		{"Message Received", MessageReceived, "twitchbot.message.received"},
		{"Message Deleted", MessageDeleted, "twitchbot.message.deleted"},
		{"Message Timeout", MessageTimeout, "twitchbot.message.timeout"},
		{"Viewer Joined", ViewerJoined, "twitchbot.viewer.joined"},
		{"Viewer Left", ViewerLeft, "twitchbot.viewer.left"},
		{"Viewer Banned", ViewerBanned, "twitchbot.viewer.banned"},
		{"Viewer Unbanned", ViewerUnbanned, "twitchbot.viewer.unbanned"},
		{"Viewer Mod Added", ViewerModAdded, "twitchbot.viewer.mod.added"},
		{"Viewer Mod Removed", ViewerModRemoved, "twitchbot.viewer.mod.removed"},
		{"Viewer VIP Added", ViewerVIPAdded, "twitchbot.viewer.vip.added"},
		{"Viewer VIP Removed", ViewerVIPRemoved, "twitchbot.viewer.vip.removed"},
		{"Clip Created", ClipCreated, "twitchbot.clip.created"},
		{"Clip Updated", ClipUpdated, "twitchbot.clip.updated"},
		{"Clip Deleted", ClipDeleted, "twitchbot.clip.deleted"},
		{"Command Created", CommandCreated, "twitchbot.command.created"},
		{"Command Updated", CommandUpdated, "twitchbot.command.updated"},
		{"Command Deleted", CommandDeleted, "twitchbot.command.deleted"},
		{"Command Executed", CommandExecuted, "twitchbot.command.executed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.event) != tt.expected {
				t.Errorf("EventType = %v, want %v", tt.event, tt.expected)
			}
		})
	}
}

func TestStreamEvent(t *testing.T) {
	event := StreamEvent{
		BaseEvent: BaseEvent{
			Type:      StreamStarted,
			Timestamp: time.Now(),
		},
		StreamID:    "test-stream",
		Title:       "Test Stream",
		GameName:    "Just Chatting",
		ViewerCount: 42,
		IsActive:    true,
	}

	if event.Type != StreamStarted {
		t.Errorf("Type = %v, want %v", event.Type, StreamStarted)
	}

	if event.StreamID != "test-stream" {
		t.Errorf("StreamID = %v, want %v", event.StreamID, "test-stream")
	}

	if event.ViewerCount != 42 {
		t.Errorf("ViewerCount = %v, want %v", event.ViewerCount, 42)
	}
}

func TestMessageEvent(t *testing.T) {
	event := MessageEvent{
		BaseEvent: BaseEvent{
			Type:      MessageReceived,
			Timestamp: time.Now(),
		},
		MessageID:    "msg-123",
		StreamID:     "stream-456",
		UserID:       "user-789",
		Username:     "testuser",
		DisplayName:  "TestUser",
		Message:      "Hello, world!",
		IsModerator:  false,
		IsSubscriber: true,
	}

	if event.Type != MessageReceived {
		t.Errorf("Type = %v, want %v", event.Type, MessageReceived)
	}

	if event.Username != "testuser" {
		t.Errorf("Username = %v, want %v", event.Username, "testuser")
	}

	if !event.IsSubscriber {
		t.Error("IsSubscriber should be true")
	}

	if event.IsModerator {
		t.Error("IsModerator should be false")
	}
}

func TestCommandEvent(t *testing.T) {
	event := CommandEvent{
		BaseEvent: BaseEvent{
			Type:      CommandExecuted,
			Timestamp: time.Now(),
		},
		CommandID:   "cmd-123",
		CommandName: "!hello",
		UserID:      "user-456",
		Username:    "testuser",
		Response:    "Hello, testuser!",
		Success:     true,
	}

	if event.Type != CommandExecuted {
		t.Errorf("Type = %v, want %v", event.Type, CommandExecuted)
	}

	if event.CommandName != "!hello" {
		t.Errorf("CommandName = %v, want %v", event.CommandName, "!hello")
	}

	if !event.Success {
		t.Error("Success should be true")
	}
}

func TestNewEventPublisher(t *testing.T) {
	publisher := NewEventPublisher(nil)

	if publisher == nil {
		t.Fatal("NewEventPublisher returned nil")
	}

	if publisher.producer != nil {
		t.Error("Producer should be nil")
	}

	// Test a few granular topics
	expectedTopics := map[EventType]string{
		StreamStarted:   "twitchbot.stream.started",
		StreamEnded:     "twitchbot.stream.ended",
		MessageReceived: "twitchbot.message.received",
		ViewerJoined:    "twitchbot.viewer.joined",
		ClipCreated:     "twitchbot.clip.created",
		CommandExecuted: "twitchbot.command.executed",
	}

	for eventType, expectedTopic := range expectedTopics {
		if publisher.topics[eventType] != expectedTopic {
			t.Errorf("Topic for %v = %v, want %v", eventType, publisher.topics[eventType], expectedTopic)
		}
	}
}

func TestPublishEventsWithoutProducer(t *testing.T) {
	publisher := NewEventPublisher(nil)

	// These should not error even with nil producer
	err := publisher.PublishStreamEvent(StreamEvent{})
	if err != nil {
		t.Errorf("PublishStreamEvent with nil producer should not error, got: %v", err)
	}

	err = publisher.PublishMessageEvent(MessageEvent{})
	if err != nil {
		t.Errorf("PublishMessageEvent with nil producer should not error, got: %v", err)
	}

	err = publisher.PublishViewerEvent(ViewerEvent{})
	if err != nil {
		t.Errorf("PublishViewerEvent with nil producer should not error, got: %v", err)
	}

	err = publisher.PublishClipEvent(ClipEvent{})
	if err != nil {
		t.Errorf("PublishClipEvent with nil producer should not error, got: %v", err)
	}

	err = publisher.PublishCommandEvent(CommandEvent{})
	if err != nil {
		t.Errorf("PublishCommandEvent with nil producer should not error, got: %v", err)
	}
}
