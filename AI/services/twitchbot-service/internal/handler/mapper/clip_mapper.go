package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

// ClipToProto converts domain Clip to protobuf
func ClipToProto(c *domain.Clip) *pb.Clip {
	if c == nil {
		return nil
	}

	return &pb.Clip{
		Id:              c.ID,
		StreamId:        c.StreamID,
		TwitchClipId:    c.TwitchClipID,
		Title:           c.Title,
		Url:             c.URL,
		EmbedUrl:        c.EmbedURL,
		ThumbnailUrl:    c.ThumbnailURL,
		CreatorName:     c.CreatorName,
		CreatorId:       c.CreatorID,
		ViewCount:       int32(c.ViewCount),
		DurationSeconds: int32(c.DurationSeconds),
		CreatedAtTwitch: TimeToProto(c.CreatedAtTwitch),
		CreatedAt:       TimeToProto(c.CreatedAt),
		UpdatedAt:       TimeToProto(c.UpdatedAt),
		DeletedAt:       timestampOrNil(nil),
	}
}

// ClipsToProto converts slice of Clips
func ClipsToProto(clips []*domain.Clip) []*pb.Clip {
	result := make([]*pb.Clip, len(clips))
	for i, c := range clips {
		result[i] = ClipToProto(c)
	}
	return result
}
