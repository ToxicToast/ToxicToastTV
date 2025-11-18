package mapper

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetDefaultPagination returns default pagination values if not provided
func GetDefaultPagination(page, pageSize int32) (int32, int32) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}

// TimeToProto converts time.Time to protobuf Timestamp
func TimeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// ProtoToTime converts protobuf Timestamp to time.Time
func ProtoToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// timestampOrNil returns a protobuf Timestamp or nil
func timestampOrNil(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringValue returns the value of a string pointer or empty string
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// BoolValue returns the value of a bool pointer or false
func BoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
