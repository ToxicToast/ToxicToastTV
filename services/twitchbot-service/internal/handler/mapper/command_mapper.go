package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

// CommandToProto converts domain Command to protobuf
func CommandToProto(c *domain.Command) *pb.Command {
	if c == nil {
		return nil
	}

	return &pb.Command{
		Id:              c.ID,
		Name:            c.Name,
		Description:     c.Description,
		Response:        c.Response,
		IsActive:        c.IsActive,
		ModeratorOnly:   c.ModeratorOnly,
		SubscriberOnly:  c.SubscriberOnly,
		CooldownSeconds: int32(c.CooldownSeconds),
		UsageCount:      int32(c.UsageCount),
		LastUsed:        TimePointerToProto(c.LastUsed),
		CreatedAt:       TimeToProto(c.CreatedAt),
		UpdatedAt:       TimeToProto(c.UpdatedAt),
		DeletedAt:       timestampOrNil(nil),
	}
}

// CommandsToProto converts slice of Commands
func CommandsToProto(commands []*domain.Command) []*pb.Command {
	result := make([]*pb.Command, len(commands))
	for i, c := range commands {
		result[i] = CommandToProto(c)
	}
	return result
}
