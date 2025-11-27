package domain

import (
	"testing"
	"time"
)

func TestStreamCreation(t *testing.T) {
	now := time.Now()
	stream := Stream{
		ID:             "test-id",
		Title:          "Test Stream",
		GameName:       "Just Chatting",
		GameID:         "509658",
		StartedAt:      now,
		PeakViewers:    100,
		AverageViewers: 75,
		TotalMessages:  500,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if stream.Title != "Test Stream" {
		t.Errorf("Title = %v, want %v", stream.Title, "Test Stream")
	}

	if stream.GameName != "Just Chatting" {
		t.Errorf("GameName = %v, want %v", stream.GameName, "Just Chatting")
	}

	if !stream.IsActive {
		t.Error("IsActive should be true")
	}

	if stream.PeakViewers != 100 {
		t.Errorf("PeakViewers = %v, want %v", stream.PeakViewers, 100)
	}
}
