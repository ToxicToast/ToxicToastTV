package grpc

import (
	"context"

	pb "toxictoast/services/warcraft-service/api/proto"
	"toxictoast/services/warcraft-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CharacterHandler struct {
	pb.UnimplementedCharacterServiceServer
	usecase *usecase.CharacterUseCase
}

func NewCharacterHandler(uc *usecase.CharacterUseCase) *CharacterHandler {
	return &CharacterHandler{usecase: uc}
}

func (h *CharacterHandler) CreateCharacter(ctx context.Context, req *pb.CreateCharacterRequest) (*pb.CharacterResponse, error) {
	character, err := h.usecase.CreateCharacter(ctx, req.Name, req.Realm, req.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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
	character, err := h.usecase.GetCharacter(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

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

	characters, total, err := h.usecase.ListCharacters(ctx, page, pageSize, req.Region, req.Realm, req.Faction)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbCharacters := make([]*pb.Character, len(characters))
	for i, char := range characters {
		pbCharacters[i] = &pb.Character{
			Id:     char.ID,
			Name:   char.Name,
			Realm:  char.Realm,
			Region: char.Region,
		}
	}

	return &pb.ListCharactersResponse{
		Characters: pbCharacters,
		Total:      int32(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

func (h *CharacterHandler) UpdateCharacter(ctx context.Context, req *pb.UpdateCharacterRequest) (*pb.CharacterResponse, error) {
	character, err := h.usecase.UpdateCharacter(ctx, req.Id, req.GuildId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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
	if err := h.usecase.DeleteCharacter(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteCharacterResponse{
		Success: true,
		Message: "Character deleted successfully",
	}, nil
}

func (h *CharacterHandler) RefreshCharacter(ctx context.Context, req *pb.RefreshCharacterRequest) (*pb.CharacterResponse, error) {
	character, err := h.usecase.RefreshCharacter(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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
	usecase *usecase.GuildUseCase
}

func NewGuildHandler(uc *usecase.GuildUseCase) *GuildHandler {
	return &GuildHandler{usecase: uc}
}

// Stub implementations - similar pattern as Character
func (h *GuildHandler) CreateGuild(ctx context.Context, req *pb.CreateGuildRequest) (*pb.GuildResponse, error) {
	guild, err := h.usecase.CreateGuild(ctx, req.Name, req.Realm, req.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

func (h *GuildHandler) ListGuilds(ctx context.Context, req *pb.ListGuildsRequest) (*pb.ListGuildsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

func (h *GuildHandler) UpdateGuild(ctx context.Context, req *pb.UpdateGuildRequest) (*pb.GuildResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

func (h *GuildHandler) DeleteGuild(ctx context.Context, req *pb.DeleteGuildRequest) (*pb.DeleteGuildResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

func (h *GuildHandler) RefreshGuild(ctx context.Context, req *pb.RefreshGuildRequest) (*pb.GuildResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

func (h *GuildHandler) GetGuildRoster(ctx context.Context, req *pb.GetGuildRosterRequest) (*pb.GuildRosterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}
