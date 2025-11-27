package grpc

import (
	"context"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	pb "toxictoast/services/warcraft-service/api/proto"
	"toxictoast/services/warcraft-service/internal/command"
	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CharacterHandler struct {
	pb.UnimplementedCharacterServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewCharacterHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *CharacterHandler {
	return &CharacterHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *CharacterHandler) CreateCharacter(ctx context.Context, req *pb.CreateCharacterRequest) (*pb.CharacterResponse, error) {
	// Create command
	cmd := &command.CreateCharacterCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		Realm:       req.Realm,
		Region:      req.Region,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create character: %v", err)
	}

	// Query created character
	getQuery := &query.GetCharacterQuery{
		BaseQuery:   cqrs.BaseQuery{},
		CharacterID: cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created character: %v", err)
	}

	character := result.(*domain.Character)

	return &pb.CharacterResponse{
		Character: &pb.Character{
			Id:     character.ID,
			Name:   character.Name,
			Realm:  character.Realm,
			Region: character.Region,
		},
	}, nil
}

func (h *CharacterHandler) GetCharacter(ctx context.Context, req *pb.GetCharacterRequest) (*pb.CharacterResponse, error) {
	// Query character
	getQuery := &query.GetCharacterQuery{
		BaseQuery:   cqrs.BaseQuery{},
		CharacterID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "character not found: %v", err)
	}

	character := result.(*domain.Character)

	return &pb.CharacterResponse{
		Character: &pb.Character{
			Id:     character.ID,
			Name:   character.Name,
			Realm:  character.Realm,
			Region: character.Region,
		},
	}, nil
}

func (h *CharacterHandler) ListCharacters(ctx context.Context, req *pb.ListCharactersRequest) (*pb.ListCharactersResponse, error) {
	page := int(req.Page)
	if page == 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 20
	}

	// Query characters
	listQuery := &query.ListCharactersQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      page,
		PageSize:  pageSize,
		Region:    req.Region,
		Realm:     req.Realm,
		Faction:   req.Faction,
	}

	result, err := h.queryBus.Dispatch(ctx, listQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list characters: %v", err)
	}

	listResult := result.(*query.ListCharactersResult)

	pbCharacters := make([]*pb.Character, len(listResult.Characters))
	for i, char := range listResult.Characters {
		pbCharacters[i] = &pb.Character{
			Id:     char.ID,
			Name:   char.Name,
			Realm:  char.Realm,
			Region: char.Region,
		}
	}

	return &pb.ListCharactersResponse{
		Characters: pbCharacters,
		Total:      int32(listResult.Total),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

func (h *CharacterHandler) UpdateCharacter(ctx context.Context, req *pb.UpdateCharacterRequest) (*pb.CharacterResponse, error) {
	// Create command
	cmd := &command.UpdateCharacterCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		GuildID:     req.GuildId,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update character: %v", err)
	}

	// Query updated character
	getQuery := &query.GetCharacterQuery{
		BaseQuery:   cqrs.BaseQuery{},
		CharacterID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve updated character: %v", err)
	}

	character := result.(*domain.Character)

	return &pb.CharacterResponse{
		Character: &pb.Character{
			Id:     character.ID,
			Name:   character.Name,
			Realm:  character.Realm,
			Region: character.Region,
		},
	}, nil
}

func (h *CharacterHandler) DeleteCharacter(ctx context.Context, req *pb.DeleteCharacterRequest) (*pb.DeleteCharacterResponse, error) {
	// Create command
	cmd := &command.DeleteCharacterCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete character: %v", err)
	}

	return &pb.DeleteCharacterResponse{
		Success: true,
		Message: "Character deleted successfully",
	}, nil
}

func (h *CharacterHandler) RefreshCharacter(ctx context.Context, req *pb.RefreshCharacterRequest) (*pb.CharacterResponse, error) {
	// Create command
	cmd := &command.RefreshCharacterCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refresh character: %v", err)
	}

	// Query refreshed character
	getQuery := &query.GetCharacterQuery{
		BaseQuery:   cqrs.BaseQuery{},
		CharacterID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve refreshed character: %v", err)
	}

	character := result.(*domain.Character)

	return &pb.CharacterResponse{
		Character: &pb.Character{
			Id:     character.ID,
			Name:   character.Name,
			Realm:  character.Realm,
			Region: character.Region,
		},
	}, nil
}

func (h *CharacterHandler) GetCharacterEquipment(ctx context.Context, req *pb.GetCharacterEquipmentRequest) (*pb.CharacterEquipmentResponse, error) {
	return nil, status.Error(codes.Unimplemented, "equipment not yet implemented")
}

func (h *CharacterHandler) GetCharacterStats(ctx context.Context, req *pb.GetCharacterStatsRequest) (*pb.CharacterStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stats not yet implemented")
}

type GuildHandler struct {
	pb.UnimplementedGuildServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewGuildHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *GuildHandler {
	return &GuildHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *GuildHandler) CreateGuild(ctx context.Context, req *pb.CreateGuildRequest) (*pb.GuildResponse, error) {
	// Create command
	cmd := &command.CreateGuildCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		Realm:       req.Realm,
		Region:      req.Region,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create guild: %v", err)
	}

	// Query created guild
	getQuery := &query.GetGuildQuery{
		BaseQuery: cqrs.BaseQuery{},
		GuildID:   cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created guild: %v", err)
	}

	guild := result.(*domain.Guild)

	return &pb.GuildResponse{
		Guild: &pb.Guild{
			Id:     guild.ID,
			Name:   guild.Name,
			Realm:  guild.Realm,
			Region: guild.Region,
		},
	}, nil
}

func (h *GuildHandler) GetGuild(ctx context.Context, req *pb.GetGuildRequest) (*pb.GuildResponse, error) {
	// Query guild
	getQuery := &query.GetGuildQuery{
		BaseQuery: cqrs.BaseQuery{},
		GuildID:   req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "guild not found: %v", err)
	}

	guild := result.(*domain.Guild)

	return &pb.GuildResponse{
		Guild: &pb.Guild{
			Id:     guild.ID,
			Name:   guild.Name,
			Realm:  guild.Realm,
			Region: guild.Region,
		},
	}, nil
}

func (h *GuildHandler) ListGuilds(ctx context.Context, req *pb.ListGuildsRequest) (*pb.ListGuildsResponse, error) {
	page := int(req.Page)
	if page == 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 20
	}

	// Query guilds
	listQuery := &query.ListGuildsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      page,
		PageSize:  pageSize,
		Region:    req.Region,
		Realm:     req.Realm,
		Faction:   req.Faction,
	}

	result, err := h.queryBus.Dispatch(ctx, listQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list guilds: %v", err)
	}

	listResult := result.(*query.ListGuildsResult)

	pbGuilds := make([]*pb.Guild, len(listResult.Guilds))
	for i, guild := range listResult.Guilds {
		pbGuilds[i] = &pb.Guild{
			Id:     guild.ID,
			Name:   guild.Name,
			Realm:  guild.Realm,
			Region: guild.Region,
		}
	}

	return &pb.ListGuildsResponse{
		Guilds:   pbGuilds,
		Total:    int32(listResult.Total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (h *GuildHandler) UpdateGuild(ctx context.Context, req *pb.UpdateGuildRequest) (*pb.GuildResponse, error) {
	// Create command
	cmd := &command.UpdateGuildCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update guild: %v", err)
	}

	// Query updated guild
	getQuery := &query.GetGuildQuery{
		BaseQuery: cqrs.BaseQuery{},
		GuildID:   req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve updated guild: %v", err)
	}

	guild := result.(*domain.Guild)

	return &pb.GuildResponse{
		Guild: &pb.Guild{
			Id:     guild.ID,
			Name:   guild.Name,
			Realm:  guild.Realm,
			Region: guild.Region,
		},
	}, nil
}

func (h *GuildHandler) DeleteGuild(ctx context.Context, req *pb.DeleteGuildRequest) (*pb.DeleteGuildResponse, error) {
	// Create command
	cmd := &command.DeleteGuildCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete guild: %v", err)
	}

	return &pb.DeleteGuildResponse{
		Success: true,
		Message: "Guild deleted successfully",
	}, nil
}

func (h *GuildHandler) RefreshGuild(ctx context.Context, req *pb.RefreshGuildRequest) (*pb.GuildResponse, error) {
	// Create command
	cmd := &command.RefreshGuildCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refresh guild: %v", err)
	}

	// Query refreshed guild
	getQuery := &query.GetGuildQuery{
		BaseQuery: cqrs.BaseQuery{},
		GuildID:   req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve refreshed guild: %v", err)
	}

	guild := result.(*domain.Guild)

	return &pb.GuildResponse{
		Guild: &pb.Guild{
			Id:     guild.ID,
			Name:   guild.Name,
			Realm:  guild.Realm,
			Region: guild.Region,
		},
	}, nil
}

func (h *GuildHandler) GetGuildRoster(ctx context.Context, req *pb.GetGuildRosterRequest) (*pb.GuildRosterResponse, error) {
	page := int(req.Page)
	if page == 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 20
	}

	// Query guild roster
	rosterQuery := &query.GetGuildRosterQuery{
		BaseQuery: cqrs.BaseQuery{},
		GuildID:   req.GuildId,
		Page:      page,
		PageSize:  pageSize,
	}

	result, err := h.queryBus.Dispatch(ctx, rosterQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get guild roster: %v", err)
	}

	rosterResult := result.(*query.GetGuildRosterResult)

	pbMembers := make([]*pb.GuildMember, len(rosterResult.Members))
	for i, member := range rosterResult.Members {
		pbMembers[i] = &pb.GuildMember{
			CharacterName: member.CharacterName,
			Rank:          int32(member.Rank),
			Level:         int32(member.Level),
		}
	}

	return &pb.GuildRosterResponse{
		Members:  pbMembers,
		Total:    int32(rosterResult.Total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
