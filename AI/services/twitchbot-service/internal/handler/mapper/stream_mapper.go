package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

// StreamToProto converts domain Stream to protobuf
func StreamToProto(s *domain.Stream) *pb.Stream {
	if s == nil {
		return nil
	}

	return &pb.Stream{
		Id:             s.ID,
		Title:          s.Title,
		GameName:       s.GameName,
		GameId:         s.GameID,
		StartedAt:      TimeToProto(s.StartedAt),
		EndedAt:        TimePointerToProto(s.EndedAt),
		PeakViewers:    int32(s.PeakViewers),
		AverageViewers: int32(s.AverageViewers),
		TotalMessages:  int32(s.TotalMessages),
		IsActive:       s.IsActive,
		CreatedAt:      TimeToProto(s.CreatedAt),
		UpdatedAt:      TimeToProto(s.UpdatedAt),
		DeletedAt:      timestampOrNil(nil),
	}
}

// StreamsToProto converts slice of Streams
func StreamsToProto(streams []*domain.Stream) []*pb.Stream {
	result := make([]*pb.Stream, len(streams))
	for i, s := range streams {
		result[i] = StreamToProto(s)
	}
	return result
}
