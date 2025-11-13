package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "toxictoast/services/warcraft-service/api/proto"
	grpcHandler "toxictoast/services/warcraft-service/internal/handler/grpc"
	"toxictoast/services/warcraft-service/internal/repository/entity"
	"toxictoast/services/warcraft-service/internal/repository/impl"
	"toxictoast/services/warcraft-service/internal/usecase"
	"toxictoast/services/warcraft-service/pkg/blizzard"
	"toxictoast/services/warcraft-service/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate
	if err := db.AutoMigrate(
		&entity.Faction{},
		&entity.Race{},
		&entity.Class{},
		&entity.Character{},
		&entity.CharacterDetails{},
		&entity.Guild{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Initialize Blizzard API client
	blizzardClient := blizzard.NewClient(
		cfg.BlizzardClientID,
		cfg.BlizzardClientSecret,
		cfg.BlizzardRegion,
	)

	// Initialize repositories
	characterRepo := impl.NewCharacterRepository(db)
	characterDetailsRepo := impl.NewCharacterDetailsRepository(db)
	guildRepo := impl.NewGuildRepository(db)

	_ = characterDetailsRepo // Will be used when details endpoints are implemented

	// Initialize use cases
	characterUseCase := usecase.NewCharacterUseCase(characterRepo, blizzardClient)
	guildUseCase := usecase.NewGuildUseCase(guildRepo, blizzardClient)

	// Initialize gRPC handlers
	characterHandler := grpcHandler.NewCharacterHandler(characterUseCase)
	guildHandler := grpcHandler.NewGuildHandler(guildUseCase)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterCharacterServiceServer(grpcServer, characterHandler)
	pb.RegisterGuildServiceServer(grpcServer, guildHandler)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Warcraft Service starting on port %s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
