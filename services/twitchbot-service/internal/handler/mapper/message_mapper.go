package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

// MessageToProto converts domain Message to protobuf
func MessageToProto(m *domain.Message) *pb.Message {
	if m == nil {
		return nil
	}

	return &pb.Message{
		Id:            m.ID,
		StreamId:      m.StreamID,
		UserId:        m.UserID,
		Username:      m.Username,
		DisplayName:   m.DisplayName,
		Message:       m.Message,
		IsModerator:   m.IsModerator,
		IsSubscriber:  m.IsSubscriber,
		IsVip:         m.IsVIP,
		IsBroadcaster: m.IsBroadcaster,
		SentAt:        TimeToProto(m.SentAt),
		CreatedAt:     TimeToProto(m.CreatedAt),
		DeletedAt:     timestampOrNil(nil),
	}
}

// MessagesToProto converts slice of Messages
func MessagesToProto(messages []*domain.Message) []*pb.Message {
	result := make([]*pb.Message, len(messages))
	for i, m := range messages {
		result[i] = MessageToProto(m)
	}
	return result
}
