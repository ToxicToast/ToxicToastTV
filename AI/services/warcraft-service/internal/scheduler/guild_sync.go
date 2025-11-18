package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"toxictoast/services/warcraft-service/internal/usecase"
)

type GuildSyncScheduler struct {
	guildUseCase *usecase.GuildUseCase
	interval     time.Duration
	enabled      bool
	stopChan     chan struct{}
}

func NewGuildSyncScheduler(
	guildUseCase *usecase.GuildUseCase,
	interval time.Duration,
	enabled bool,
) *GuildSyncScheduler {
	return &GuildSyncScheduler{
		guildUseCase: guildUseCase,
		interval:     interval,
		enabled:      enabled,
		stopChan:     make(chan struct{}),
	}
}

func (s *GuildSyncScheduler) Start() {
	if !s.enabled {
		log.Println("Guild sync scheduler is disabled")
		return
	}

	log.Printf("Guild sync scheduler started (interval: %v)", s.interval)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.syncAllGuilds()
			case <-s.stopChan:
				log.Println("Guild sync scheduler stopped")
				return
			}
		}
	}()
}

func (s *GuildSyncScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *GuildSyncScheduler) syncAllGuilds() {
	ctx := context.Background()
	log.Println("Starting scheduled guild sync...")

	// Get all guilds (first 1000, can be adjusted)
	guilds, total, err := s.guildUseCase.ListGuilds(ctx, 1, 1000, nil, nil, nil)
	if err != nil {
		log.Printf("Error listing guilds for sync: %v", err)
		return
	}

	log.Printf("Found %d guilds to sync", total)

	successCount := 0
	errorCount := 0

	for _, guild := range guilds {
		// Refresh guild data from Blizzard API
		_, err := s.guildUseCase.RefreshGuild(ctx, guild.ID)
		if err != nil {
			log.Printf("Error syncing guild %s-%s: %v", guild.Name, guild.Realm, err)
			errorCount++
		} else {
			log.Printf("Successfully synced guild %s-%s", guild.Name, guild.Realm)
			successCount++
		}

		// Small delay between API calls to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Guild sync completed: %d successful, %d errors", successCount, errorCount)

	if successCount > 0 || errorCount > 0 {
		fmt.Printf("\nScheduled guild sync completed: %d/%d guilds synced successfully\n", successCount, successCount+errorCount)
	}
}
