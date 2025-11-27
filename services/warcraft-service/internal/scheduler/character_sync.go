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

type CharacterSyncScheduler struct {
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
	interval   time.Duration
	enabled    bool
	stopChan   chan struct{}
}

func NewCharacterSyncScheduler(
	commandBus *cqrs.CommandBus,
	queryBus *cqrs.QueryBus,
	interval time.Duration,
	enabled bool,
) *CharacterSyncScheduler {
	return &CharacterSyncScheduler{
		commandBus: commandBus,
		queryBus:   queryBus,
		interval:   interval,
		enabled:    enabled,
		stopChan:   make(chan struct{}),
	}
}

func (s *CharacterSyncScheduler) Start() {
	if !s.enabled {
		log.Println("Character sync scheduler is disabled")
		return
	}

	log.Printf("Character sync scheduler started (interval: %v)", s.interval)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.syncAllCharacters()
			case <-s.stopChan:
				log.Println("Character sync scheduler stopped")
				return
			}
		}
	}()
}

func (s *CharacterSyncScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *CharacterSyncScheduler) syncAllCharacters() {
	ctx := context.Background()
	log.Println("Starting scheduled character sync...")

	// Get all characters (first 1000, can be adjusted)
	listQuery := &query.ListCharactersQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      1,
		PageSize:  1000,
	}

	result, err := s.queryBus.Dispatch(ctx, listQuery)
	if err != nil {
		log.Printf("Error listing characters for sync: %v", err)
		return
	}

	listResult := result.(*query.ListCharactersResult)
	log.Printf("Found %d characters to sync", listResult.Total)

	successCount := 0
	errorCount := 0

	for _, character := range listResult.Characters {
		// Refresh character data from Blizzard API
		refreshCmd := &command.RefreshCharacterCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: character.ID},
		}

		err := s.commandBus.Dispatch(ctx, refreshCmd)
		if err != nil {
			log.Printf("Error syncing character %s-%s: %v", character.Name, character.Realm, err)
			errorCount++
		} else {
			log.Printf("Successfully synced character %s-%s", character.Name, character.Realm)
			successCount++
		}

		// Small delay between API calls to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Character sync completed: %d successful, %d errors", successCount, errorCount)

	if successCount > 0 || errorCount > 0 {
		fmt.Printf("\nScheduled character sync completed: %d/%d characters synced successfully\n", successCount, successCount+errorCount)
	}
}
