package impl

import (
	"log"
	"time"

	"gorm.io/gorm"

	"toxictoast/services/twitchbot-service/internal/domain"
)

const chatOnlyStreamID = "00000000-0000-0000-0000-000000000001"

// EnsureChatOnlyStream ensures the Chat-Only stream exists in the database
// This stream is used for logging messages when no active stream is running
func EnsureChatOnlyStream(db *gorm.DB) error {
	log.Println("Ensuring Chat-Only stream exists...")

	// Check if Chat-Only stream already exists
	var existing domain.Stream
	result := db.First(&existing, "id = ?", chatOnlyStreamID)

	if result.Error == nil {
		log.Println("✅ Chat-Only stream already exists")
		return nil
	}

	// Create the Chat-Only stream
	now := time.Now()
	chatOnlyStream := &domain.Stream{
		ID:             chatOnlyStreamID,
		Title:          "Chat-Only Messages",
		GameName:       "Just Chatting",
		GameID:         "509658", // Twitch Game ID for "Just Chatting"
		StartedAt:      now,
		EndedAt:        nil,
		PeakViewers:    0,
		AverageViewers: 0,
		TotalMessages:  0,
		IsActive:       true,
	}

	if err := db.Create(chatOnlyStream).Error; err != nil {
		log.Printf("❌ Failed to create Chat-Only stream: %v", err)
		return err
	}

	log.Printf("✅ Chat-Only stream created successfully (ID: %s)", chatOnlyStreamID)
	return nil
}
