package mapper

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// timestampOrNil converts *time.Time to protobuf timestamp, returning nil if input is nil
func timestampOrNil(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
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
