package mapper

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"toxictoast/services/sse-service/internal/domain"
	pb "toxictoast/services/sse-service/api/proto"
)

// ClientStatsToProto converts domain ClientStats to protobuf
func ClientStatsToProto(stats domain.ClientStats) *pb.ClientInfo {
	return &pb.ClientInfo{
		Id:             stats.ID,
		ConnectedAt:    timestamppb.New(stats.ConnectedAt),
		LastEventAt:    timestamppb.New(stats.LastEventAt),
		EventsReceived: stats.EventsReceived,
		Filter:         SubscriptionFilterToProto(stats.Filter),
		UserAgent:      stats.UserAgent,
		RemoteAddr:     stats.RemoteAddr,
	}
}

// ClientStatsListToProto converts slice of ClientStats to protobuf
func ClientStatsListToProto(statsList []domain.ClientStats) []*pb.ClientInfo {
	result := make([]*pb.ClientInfo, len(statsList))
	for i, stats := range statsList {
		result[i] = ClientStatsToProto(stats)
	}
	return result
}

// SubscriptionFilterToProto converts domain SubscriptionFilter to protobuf
func SubscriptionFilterToProto(filter domain.SubscriptionFilter) *pb.SubscriptionFilter {
	eventTypes := make([]string, len(filter.EventTypes))
	for i, et := range filter.EventTypes {
		eventTypes[i] = string(et)
	}

	return &pb.SubscriptionFilter{
		EventTypes: eventTypes,
		Sources:    filter.Sources,
	}
}
