package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/toxictoast/toxictoastgo/shared/config"
)

// Connect connects to PostgreSQL with retry logic
func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDatabaseURL()

	log.Printf("Connecting to database at %s:%s/%s", cfg.Host, cfg.Port, cfg.Name)

	// Retry connection up to 5 times
	for i := 0; i < 5; i++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("Attempt %d: Failed to open database connection: %v", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Get underlying sql.DB
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Attempt %d: Failed to get underlying SQL DB: %v", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Configure connection pool
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = sqlDB.PingContext(ctx)
		cancel()

		if err == nil {
			return db, nil
		}

		log.Printf("Attempt %d: Database ping failed: %v", i+1, err)
		sqlDBClose, _ := db.DB()
		if sqlDBClose != nil {
			sqlDBClose.Close()
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after 5 attempts")
}

// CheckHealth checks if database is healthy
func CheckHealth(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}

// AutoMigrate runs GORM auto-migrations for provided entities
func AutoMigrate(db *gorm.DB, entities ...interface{}) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	log.Println("Running GORM auto-migrations...")

	for _, entity := range entities {
		if err := db.AutoMigrate(entity); err != nil {
			return fmt.Errorf("failed to auto-migrate entity %T: %w", entity, err)
		}
		log.Printf("Migrated entity: %T", entity)
	}

	log.Println("GORM auto-migrations completed successfully")
	return nil
}
