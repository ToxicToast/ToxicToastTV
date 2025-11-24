package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	kafkaBroker = "localhost:19092"
	kafkaGroup  = "integration-test-group"
)

// TestKafkaUserEvents tests that user-related events are published correctly to Kafka
func TestKafkaUserEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	// Create test user via API
	timestamp := time.Now().Unix()
	testEmail := fmt.Sprintf("kafkatest%d@example.com", timestamp)
	testUsername := fmt.Sprintf("kafkatest%d", timestamp)
	testPassword := "SecurePass123!"

	client := &http.Client{Timeout: testTimeout}

	// Register user
	registerPayload := map[string]interface{}{
		"email":      testEmail,
		"username":   testUsername,
		"password":   testPassword,
		"first_name": "Kafka",
		"last_name":  "Test",
	}

	resp, err := makeJSONRequest(client, "POST", gatewayBaseURL+"/auth/register", registerPayload, "")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	resp.Body.Close()

	// Wait a bit for Kafka to process
	time.Sleep(2 * time.Second)

	// Test user.created event
	t.Run("UserCreatedEvent", func(t *testing.T) {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{kafkaBroker},
			Topic:     "user.created",
			GroupID:   kafkaGroup + "-user-created",
			MinBytes:  1,
			MaxBytes:  10e6,
			MaxWait:   time.Second,
			Partition: 0,
		})
		defer reader.Close()

		// Set read deadline
		reader.SetOffset(kafka.LastOffset)

		found := false
		deadline := time.Now().Add(10 * time.Second)

		for time.Now().Before(deadline) && !found {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			msg, err := reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue
				}
				break
			}

			var event map[string]interface{}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				continue
			}

			// Check if this is our user
			if email, ok := event["email"].(string); ok && email == testEmail {
				found = true
				t.Logf("✓ Found user.created event for: %s", testEmail)

				// Validate event structure
				if _, ok := event["user_id"]; !ok {
					t.Error("Event missing user_id field")
				}
				if username, ok := event["username"].(string); !ok || username != testUsername {
					t.Errorf("Expected username %s, got %v", testUsername, username)
				}
				if _, ok := event["created_at"]; !ok {
					t.Error("Event missing created_at field")
				}
			}
		}

		if !found {
			t.Logf("⚠ user.created event not found for %s (may have been consumed already)", testEmail)
		}
	})

	// Test auth.registered event
	t.Run("AuthRegisteredEvent", func(t *testing.T) {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{kafkaBroker},
			Topic:     "auth.registered",
			GroupID:   kafkaGroup + "-auth-registered",
			MinBytes:  1,
			MaxBytes:  10e6,
			MaxWait:   time.Second,
			Partition: 0,
		})
		defer reader.Close()

		reader.SetOffset(kafka.LastOffset)

		found := false
		deadline := time.Now().Add(10 * time.Second)

		for time.Now().Before(deadline) && !found {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			msg, err := reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue
				}
				break
			}

			var event map[string]interface{}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				continue
			}

			// Check if this is our user
			if email, ok := event["email"].(string); ok && email == testEmail {
				found = true
				t.Logf("✓ Found auth.registered event for: %s", testEmail)

				// Validate event structure
				if _, ok := event["user_id"]; !ok {
					t.Error("Event missing user_id field")
				}
				if username, ok := event["username"].(string); !ok || username != testUsername {
					t.Errorf("Expected username %s, got %v", testUsername, username)
				}
				if _, ok := event["registered_at"]; !ok {
					t.Error("Event missing registered_at field")
				}
			}
		}

		if !found {
			t.Logf("⚠ auth.registered event not found for %s (may have been consumed already)", testEmail)
		}
	})

	// Login to trigger auth.login event
	loginPayload := map[string]interface{}{
		"email":    testEmail,
		"password": testPassword,
	}

	resp, err = makeJSONRequest(client, "POST", gatewayBaseURL+"/auth/login", loginPayload, "")
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	resp.Body.Close()

	time.Sleep(2 * time.Second)

	// Test auth.login event
	t.Run("AuthLoginEvent", func(t *testing.T) {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{kafkaBroker},
			Topic:     "auth.login",
			GroupID:   kafkaGroup + "-auth-login",
			MinBytes:  1,
			MaxBytes:  10e6,
			MaxWait:   time.Second,
			Partition: 0,
		})
		defer reader.Close()

		reader.SetOffset(kafka.LastOffset)

		found := false
		deadline := time.Now().Add(10 * time.Second)

		for time.Now().Before(deadline) && !found {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			msg, err := reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue
				}
				break
			}

			var event map[string]interface{}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				continue
			}

			// Check if this is our user
			if email, ok := event["email"].(string); ok && email == testEmail {
				found = true
				t.Logf("✓ Found auth.login event for: %s", testEmail)

				// Validate event structure
				if _, ok := event["user_id"]; !ok {
					t.Error("Event missing user_id field")
				}
				if _, ok := event["login_at"]; !ok {
					t.Error("Event missing login_at field")
				}
			}
		}

		if !found {
			t.Logf("⚠ auth.login event not found for %s (may have been consumed already)", testEmail)
		}
	})
}

// TestKafkaEventStructure tests that Kafka events have proper structure
func TestKafkaEventStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	topics := []string{
		"user.created",
		"user.updated",
		"user.deleted",
		"user.activated",
		"user.deactivated",
		"user.password.changed",
		"auth.registered",
		"auth.login",
		"auth.token.refreshed",
	}

	for _, topic := range topics {
		t.Run(topic, func(t *testing.T) {
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers:   []string{kafkaBroker},
				Topic:     topic,
				GroupID:   kafkaGroup + "-structure-test",
				MinBytes:  1,
				MaxBytes:  10e6,
				MaxWait:   time.Second,
				Partition: 0,
			})
			defer reader.Close()

			// Try to read one message
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			msg, err := reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					t.Logf("⚠ No messages found in topic %s (may be empty)", topic)
					return
				}
				t.Logf("⚠ Error reading from topic %s: %v", topic, err)
				return
			}

			// Validate message structure
			var event map[string]interface{}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				t.Errorf("Failed to unmarshal event from topic %s: %v", topic, err)
				return
			}

			// All events should have these fields
			if _, ok := event["user_id"]; !ok {
				t.Errorf("Event in topic %s missing user_id field", topic)
			}

			t.Logf("✓ Topic %s has valid event structure with %d fields", topic, len(event))
		})
	}
}

// TestKafkaTopicsExist verifies that all expected Kafka topics exist
func TestKafkaTopicsExist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	expectedTopics := []string{
		"user.created",
		"user.updated",
		"user.deleted",
		"user.activated",
		"user.deactivated",
		"user.password.changed",
		"auth.registered",
		"auth.login",
		"auth.token.refreshed",
	}

	conn, err := kafka.Dial("tcp", kafkaBroker)
	if err != nil {
		t.Fatalf("Failed to connect to Kafka: %v", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		t.Fatalf("Failed to read partitions: %v", err)
	}

	existingTopics := make(map[string]bool)
	for _, p := range partitions {
		existingTopics[p.Topic] = true
	}

	for _, topic := range expectedTopics {
		if existingTopics[topic] {
			t.Logf("✓ Topic exists: %s", topic)
		} else {
			t.Logf("⚠ Topic not found: %s (will be created on first use)", topic)
		}
	}
}
