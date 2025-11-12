package mapper

import (
	"strings"

	"toxictoast/services/notification-service/internal/domain"
	pb "toxictoast/services/notification-service/api/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoChannel(channel *domain.DiscordChannel) *pb.DiscordChannel {
	if channel == nil {
		return nil
	}

	eventTypes := []string{}
	if channel.EventTypes != "" {
		eventTypes = strings.Split(channel.EventTypes, ",")
	}

	proto := &pb.DiscordChannel{
		Id:                   channel.ID,
		Name:                 channel.Name,
		WebhookUrl:           channel.WebhookURL,
		EventTypes:           eventTypes,
		Color:                int32(channel.Color),
		Active:               channel.Active,
		Description:          channel.Description,
		TotalNotifications:   int32(channel.TotalNotifications),
		SuccessNotifications: int32(channel.SuccessNotifications),
		FailedNotifications:  int32(channel.FailedNotifications),
		CreatedAt:            timestamppb.New(channel.CreatedAt),
		UpdatedAt:            timestamppb.New(channel.UpdatedAt),
	}

	if channel.LastNotificationAt != nil {
		proto.LastNotificationAt = timestamppb.New(*channel.LastNotificationAt)
	}
	if channel.LastSuccessAt != nil {
		proto.LastSuccessAt = timestamppb.New(*channel.LastSuccessAt)
	}
	if channel.LastFailureAt != nil {
		proto.LastFailureAt = timestamppb.New(*channel.LastFailureAt)
	}

	return proto
}

func ToProtoChannels(channels []*domain.DiscordChannel) []*pb.DiscordChannel {
	result := make([]*pb.DiscordChannel, 0, len(channels))
	for _, channel := range channels {
		result = append(result, ToProtoChannel(channel))
	}
	return result
}
