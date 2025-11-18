package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

// ChannelViewerToProto converts domain ChannelViewer to protobuf
func ChannelViewerToProto(cv *domain.ChannelViewer) *pb.ChannelViewer {
	if cv == nil {
		return nil
	}

	return &pb.ChannelViewer{
		Id:          cv.ID,
		Channel:     cv.Channel,
		TwitchId:    cv.TwitchID,
		Username:    cv.Username,
		DisplayName: cv.DisplayName,
		FirstSeen:   TimeToProto(cv.FirstSeen),
		LastSeen:    TimeToProto(cv.LastSeen),
		IsModerator: cv.IsModerator,
		IsVip:       cv.IsVIP,
		CreatedAt:   TimeToProto(cv.CreatedAt),
		UpdatedAt:   TimeToProto(cv.UpdatedAt),
	}
}

// ChannelViewersToProto converts slice of ChannelViewers
func ChannelViewersToProto(viewers []*domain.ChannelViewer) []*pb.ChannelViewer {
	result := make([]*pb.ChannelViewer, len(viewers))
	for i, v := range viewers {
		result[i] = ChannelViewerToProto(v)
	}
	return result
}
