package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/warcraft-service/internal/command"
	"toxictoast/services/warcraft-service/internal/query"
)

type GuildSyncScheduler struct {
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
	interval   time.Duration
	enabled    bool
	stopChan   chan struct{}
}

func NewGuildSyncScheduler(
	commandBus *cqrs.CommandBus,
	queryBus *cqrs.QueryBus,
	interval time.Duration,
	enabled bool,
) *GuildSyncScheduler {
	return &GuildSyncScheduler{
		commandBus: commandBus,
		queryBus:   queryBus,
		interval:   interval,
		enabled:    enabled,
		stopChan:   make(chan struct{}),
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
	listQuery := &query.ListGuildsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      1,
		PageSize:  1000,
	}

	result, err := s.queryBus.Dispatch(ctx, listQuery)
	if err != nil {
		log.Printf("Error listing guilds for sync: %v", err)
		return
	}

	listResult := result.(*query.ListGuildsResult)
	log.Printf("Found %d guilds to sync", listResult.Total)

	successCount := 0
	errorCount := 0

	for _, guild := range listResult.Guilds {
		// Refresh guild data from Blizzard API
		refreshCmd := &command.RefreshGuildCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: guild.ID},
		}

		err := s.commandBus.Dispatch(ctx, refreshCmd)
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
