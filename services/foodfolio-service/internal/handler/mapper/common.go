package mapper

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

// ToTimestamps converts domain timestamps to protobuf timestamps
func ToTimestamps(createdAt, updatedAt time.Time, deletedAt *time.Time) *pb.Timestamps {
	ts := &pb.Timestamps{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	if deletedAt != nil {
		ts.DeletedAt = timestamppb.New(*deletedAt)
	}

	return ts
}

// ToPaginationResponse converts pagination info to protobuf
func ToPaginationResponse(page, pageSize int, totalItems int64) *pb.PaginationResponse {
	totalPages := int32((totalItems + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	return &pb.PaginationResponse{
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalItems: int32(totalItems),
		TotalPages: totalPages,
	}
}

// TimeToProto converts Go time to protobuf timestamp
func TimeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// TimePointerToProto converts Go *time.Time to protobuf timestamp
func TimePointerToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// ProtoToTime converts protobuf timestamp to Go time
func ProtoToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// ProtoToTimePointer converts protobuf timestamp to Go *time.Time
func ProtoToTimePointer(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
