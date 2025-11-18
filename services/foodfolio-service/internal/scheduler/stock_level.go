package scheduler

import (
	"context"
	"log"
	"time"

	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
	"toxictoast/services/foodfolio-service/internal/usecase"
)

type StockLevelScheduler struct {
	itemVariantUseCase usecase.ItemVariantUseCase
	itemVariantRepo    interfaces.ItemVariantRepository
	interval           time.Duration
	enabled            bool
	stopChan           chan struct{}
}

func NewStockLevelScheduler(
	itemVariantUseCase usecase.ItemVariantUseCase,
	itemVariantRepo interfaces.ItemVariantRepository,
	interval time.Duration,
	enabled bool,
) *StockLevelScheduler {
	return &StockLevelScheduler{
		itemVariantUseCase: itemVariantUseCase,
		itemVariantRepo:    itemVariantRepo,
		interval:           interval,
		enabled:            enabled,
		stopChan:           make(chan struct{}),
	}
}

func (s *StockLevelScheduler) Start() {
	if !s.enabled {
		log.Println("Stock level scheduler is disabled")
		return
	}

	log.Printf("Stock level scheduler started (interval: %v)", s.interval)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.checkStockLevels()

		for {
			select {
			case <-ticker.C:
				s.checkStockLevels()
			case <-s.stopChan:
				log.Println("Stock level scheduler stopped")
				return
			}
		}
	}()
}

func (s *StockLevelScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *StockLevelScheduler) checkStockLevels() {
	ctx := context.Background()
	log.Println("Checking stock levels...")

	// Get all active item variants
	offset := 0
	limit := 1000

	variants, total, err := s.itemVariantRepo.List(ctx, offset, limit, nil, nil, nil, false)
	if err != nil {
		log.Printf("Error listing item variants for stock check: %v", err)
		return
	}

	log.Printf("Found %d item variants to check", total)

	lowStockCount := 0
	emptyStockCount := 0
	errorCount := 0

	for i := range variants {
		variant := variants[i]

		// Get current stock
		currentStock := variant.CurrentStock()

		// Check if empty
		if currentStock == 0 {
			err := s.itemVariantUseCase.NotifyStockEmpty(ctx, variant)
			if err != nil {
				log.Printf("Error notifying empty stock for variant %s: %v", variant.ID, err)
				errorCount++
				continue
			}

			log.Printf("Notified empty stock: %s (current: %d, min: %d)", variant.VariantName, currentStock, variant.MinSKU)
			emptyStockCount++
		} else if currentStock < variant.MinSKU {
			// Check if below minimum
			err := s.itemVariantUseCase.NotifyStockLow(ctx, variant, currentStock)
			if err != nil {
				log.Printf("Error notifying low stock for variant %s: %v", variant.ID, err)
				errorCount++
				continue
			}

			log.Printf("Notified low stock: %s (current: %d, min: %d)", variant.VariantName, currentStock, variant.MinSKU)
			lowStockCount++
		}

		time.Sleep(10 * time.Millisecond)
	}

	if lowStockCount > 0 || emptyStockCount > 0 || errorCount > 0 {
		log.Printf("Stock level check completed: %d low stock, %d empty stock, %d errors",
			lowStockCount, emptyStockCount, errorCount)
	}
}
