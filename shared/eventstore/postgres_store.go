package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// PostgresEventStore implements EventStore using PostgreSQL
type PostgresEventStore struct {
	db *sql.DB
}

// NewPostgresEventStore creates a new PostgreSQL event store
func NewPostgresEventStore(db *sql.DB) (*PostgresEventStore, error) {
	store := &PostgresEventStore{db: db}

	// Create events table if it doesn't exist
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create event store tables: %w", err)
	}

	return store, nil
}

// createTables creates the necessary tables for event sourcing
func (s *PostgresEventStore) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS event_store (
		event_id VARCHAR(36) PRIMARY KEY,
		event_type VARCHAR(255) NOT NULL,
		aggregate_id VARCHAR(36) NOT NULL,
		aggregate_type VARCHAR(255) NOT NULL,
		version BIGINT NOT NULL,
		timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
		data JSONB NOT NULL,
		metadata JSONB,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(aggregate_id, version)
	);

	CREATE INDEX IF NOT EXISTS idx_event_store_aggregate ON event_store(aggregate_type, aggregate_id);
	CREATE INDEX IF NOT EXISTS idx_event_store_type ON event_store(event_type);
	CREATE INDEX IF NOT EXISTS idx_event_store_timestamp ON event_store(timestamp);
	CREATE INDEX IF NOT EXISTS idx_event_store_aggregate_version ON event_store(aggregate_id, version);
	`

	_, err := s.db.Exec(query)
	return err
}

// SaveEvents saves events with optimistic locking
func (s *PostgresEventStore) SaveEvents(ctx context.Context, aggregateID string, expectedVersion int64, events []*EventEnvelope) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check current version (optimistic locking)
	var currentVersion sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT MAX(version) FROM event_store WHERE aggregate_id = $1`,
		aggregateID,
	).Scan(&currentVersion)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	actualVersion := int64(-1)
	if currentVersion.Valid {
		actualVersion = currentVersion.Int64
	}

	if actualVersion != expectedVersion {
		return ErrConcurrencyConflict
	}

	// Insert events
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO event_store (
			event_id, event_type, aggregate_id, aggregate_type,
			version, timestamp, data, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			event.EventID,
			event.EventType,
			event.AggregateID,
			event.AggregateType,
			event.Version,
			event.Timestamp,
			event.Data,
			metadataJSON,
		)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" { // unique violation
					return ErrConcurrencyConflict
				}
			}
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetEvents retrieves all events for an aggregate
func (s *PostgresEventStore) GetEvents(ctx context.Context, aggregateType, aggregateID string) ([]*EventEnvelope, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT event_id, event_type, aggregate_id, aggregate_type, version, timestamp, data, metadata
		FROM event_store
		WHERE aggregate_type = $1 AND aggregate_id = $2
		ORDER BY version ASC
	`, aggregateType, aggregateID)

	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

// GetEventsSince retrieves events since a specific version
func (s *PostgresEventStore) GetEventsSince(ctx context.Context, aggregateType, aggregateID string, sinceVersion int64) ([]*EventEnvelope, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT event_id, event_type, aggregate_id, aggregate_type, version, timestamp, data, metadata
		FROM event_store
		WHERE aggregate_type = $1 AND aggregate_id = $2 AND version > $3
		ORDER BY version ASC
	`, aggregateType, aggregateID, sinceVersion)

	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

// GetEventsByType retrieves events of a specific type
func (s *PostgresEventStore) GetEventsByType(ctx context.Context, aggregateType, aggregateID, eventType string) ([]*EventEnvelope, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT event_id, event_type, aggregate_id, aggregate_type, version, timestamp, data, metadata
		FROM event_store
		WHERE aggregate_type = $1 AND aggregate_id = $2 AND event_type = $3
		ORDER BY version ASC
	`, aggregateType, aggregateID, eventType)

	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

// GetAllEvents retrieves all events of a specific aggregate type
func (s *PostgresEventStore) GetAllEvents(ctx context.Context, aggregateType string, limit, offset int) ([]*EventEnvelope, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT event_id, event_type, aggregate_id, aggregate_type, version, timestamp, data, metadata
		FROM event_store
		WHERE aggregate_type = $1
		ORDER BY timestamp ASC, version ASC
		LIMIT $2 OFFSET $3
	`, aggregateType, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

// GetEventStream retrieves events across all aggregates
func (s *PostgresEventStore) GetEventStream(ctx context.Context, sinceTimestamp int64, limit int) ([]*EventEnvelope, error) {
	timestamp := time.Unix(sinceTimestamp, 0).UTC()

	rows, err := s.db.QueryContext(ctx, `
		SELECT event_id, event_type, aggregate_id, aggregate_type, version, timestamp, data, metadata
		FROM event_store
		WHERE timestamp > $1
		ORDER BY timestamp ASC
		LIMIT $2
	`, timestamp, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query event stream: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

// GetAggregateVersion gets the current version of an aggregate
func (s *PostgresEventStore) GetAggregateVersion(ctx context.Context, aggregateType, aggregateID string) (int64, error) {
	var version sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT MAX(version)
		FROM event_store
		WHERE aggregate_type = $1 AND aggregate_id = $2
	`, aggregateType, aggregateID).Scan(&version)

	if err != nil {
		return -1, fmt.Errorf("failed to get aggregate version: %w", err)
	}

	if !version.Valid {
		return -1, nil // Aggregate doesn't exist yet
	}

	return version.Int64, nil
}

// scanEvents scans database rows into event envelopes
func (s *PostgresEventStore) scanEvents(rows *sql.Rows) ([]*EventEnvelope, error) {
	var events []*EventEnvelope

	for rows.Next() {
		var event EventEnvelope
		var metadataJSON []byte

		err := rows.Scan(
			&event.EventID,
			&event.EventType,
			&event.AggregateID,
			&event.AggregateType,
			&event.Version,
			&event.Timestamp,
			&event.Data,
			&metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// Close closes the database connection
func (s *PostgresEventStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
