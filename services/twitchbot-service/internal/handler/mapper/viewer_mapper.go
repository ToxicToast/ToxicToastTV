package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

// ViewerToProto converts domain Viewer to protobuf
func ViewerToProto(v *domain.Viewer) *pb.Viewer {
	if v == nil {
		return nil
	}

	return &pb.Viewer{
		Id:                  v.ID,
		TwitchId:            v.TwitchID,
		Username:            v.Username,
		DisplayName:         v.DisplayName,
		TotalMessages:       int32(v.TotalMessages),
		TotalStreamsWatched: int32(v.TotalStreamsWatched),
		FirstSeen:           TimeToProto(v.FirstSeen),
		LastSeen:            TimeToProto(v.LastSeen),
		CreatedAt:           TimeToProto(v.CreatedAt),
		UpdatedAt:           TimeToProto(v.UpdatedAt),
		DeletedAt:           timestampOrNil(nil),
	}
}

// ViewersToProto converts slice of Viewers
func ViewersToProto(viewers []*domain.Viewer) []*pb.Viewer {
	result := make([]*pb.Viewer, len(viewers))
	for i, v := range viewers {
		result[i] = ViewerToProto(v)
	}
	return result
}
