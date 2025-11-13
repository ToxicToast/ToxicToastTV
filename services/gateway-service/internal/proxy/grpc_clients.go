package proxy

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceClients holds gRPC client connections to all backend services
type ServiceClients struct {
	BlogConn         *grpc.ClientConn
	LinkConn         *grpc.ClientConn
	FoodfolioConn    *grpc.ClientConn
	NotificationConn *grpc.ClientConn
	SSEConn          *grpc.ClientConn
	TwitchBotConn    *grpc.ClientConn
	WebhookConn      *grpc.ClientConn
	WarcraftConn     *grpc.ClientConn
}

// NewServiceClients creates gRPC connections to all backend services
func NewServiceClients(ctx context.Context, config ServiceURLs) (*ServiceClients, error) {
	clients := &ServiceClients{}

	var err error

	// Connect to blog service
	if config.BlogURL != "" {
		clients.BlogConn, err = grpc.NewClient(
			config.BlogURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to blog service: %w", err)
		}
	}

	// Connect to link service
	if config.LinkURL != "" {
		clients.LinkConn, err = grpc.NewClient(
			config.LinkURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to link service: %w", err)
		}
	}

	// Connect to foodfolio service
	if config.FoodfolioURL != "" {
		clients.FoodfolioConn, err = grpc.NewClient(
			config.FoodfolioURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to foodfolio service: %w", err)
		}
	}

	// Connect to notification service
	if config.NotificationURL != "" {
		clients.NotificationConn, err = grpc.NewClient(
			config.NotificationURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to notification service: %w", err)
		}
	}

	// Connect to SSE service
	if config.SSEURL != "" {
		clients.SSEConn, err = grpc.NewClient(
			config.SSEURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SSE service: %w", err)
		}
	}

	// Connect to TwitchBot service
	if config.TwitchBotURL != "" {
		clients.TwitchBotConn, err = grpc.NewClient(
			config.TwitchBotURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to twitchbot service: %w", err)
		}
	}

	// Connect to Webhook service
	if config.WebhookURL != "" {
		clients.WebhookConn, err = grpc.NewClient(
			config.WebhookURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to webhook service: %w", err)
		}
	}

	// Connect to Warcraft service
	if config.WarcraftURL != "" {
		clients.WarcraftConn, err = grpc.NewClient(
			config.WarcraftURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to warcraft service: %w", err)
		}
	}

	return clients, nil
}

// Close closes all gRPC connections
func (sc *ServiceClients) Close() error {
	if sc.BlogConn != nil {
		sc.BlogConn.Close()
	}
	if sc.LinkConn != nil {
		sc.LinkConn.Close()
	}
	if sc.FoodfolioConn != nil {
		sc.FoodfolioConn.Close()
	}
	if sc.NotificationConn != nil {
		sc.NotificationConn.Close()
	}
	if sc.SSEConn != nil {
		sc.SSEConn.Close()
	}
	if sc.TwitchBotConn != nil {
		sc.TwitchBotConn.Close()
	}
	if sc.WebhookConn != nil {
		sc.WebhookConn.Close()
	}
	if sc.WarcraftConn != nil {
		sc.WarcraftConn.Close()
	}
	return nil
}

// ServiceURLs holds the URLs for all backend services
type ServiceURLs struct {
	BlogURL         string
	LinkURL         string
	FoodfolioURL    string
	NotificationURL string
	SSEURL          string
	TwitchBotURL    string
	WebhookURL      string
	WarcraftURL     string
}
