package mapper

import (
	pb "toxictoast/services/link-service/api/proto"
	"toxictoast/services/link-service/internal/domain"
)

// LinkToProto converts a domain Link to a protobuf Link
func LinkToProto(link *domain.Link) *pb.Link {
	if link == nil {
		return nil
	}

	protoLink := &pb.Link{
		Id:          link.ID,
		OriginalUrl: link.OriginalURL,
		ShortCode:   link.ShortCode,
		IsActive:    link.IsActive,
		ClickCount:  int32(link.ClickCount),
		CreatedAt:   TimeToProto(link.CreatedAt),
		UpdatedAt:   TimeToProto(link.UpdatedAt),
	}

	// Handle optional fields
	if link.CustomAlias != nil {
		protoLink.CustomAlias = link.CustomAlias
	}

	if link.Title != nil {
		protoLink.Title = link.Title
	}

	if link.Description != nil {
		protoLink.Description = link.Description
	}

	if link.ExpiresAt != nil {
		protoLink.ExpiresAt = timestampOrNil(link.ExpiresAt)
	}

	if !link.DeletedAt.Time.IsZero() {
		t := link.DeletedAt.Time
		protoLink.DeletedAt = timestampOrNil(&t)
	}

	return protoLink
}

// LinksToProto converts a slice of domain Links to protobuf Links
func LinksToProto(links []domain.Link) []*pb.Link {
	protoLinks := make([]*pb.Link, len(links))
	for i, link := range links {
		protoLinks[i] = LinkToProto(&link)
	}
	return protoLinks
}

// ClickToProto converts a domain Click to a protobuf Click
func ClickToProto(click *domain.Click) *pb.Click {
	if click == nil {
		return nil
	}

	protoClick := &pb.Click{
		Id:        click.ID,
		LinkId:    click.LinkID,
		IpAddress: click.IPAddress,
		UserAgent: click.UserAgent,
		ClickedAt: TimeToProto(click.ClickedAt),
		CreatedAt: TimeToProto(click.CreatedAt),
	}

	// Handle optional fields
	if click.Referer != nil {
		protoClick.Referer = click.Referer
	}

	if click.Country != nil {
		protoClick.Country = click.Country
	}

	if click.City != nil {
		protoClick.City = click.City
	}

	if click.DeviceType != nil {
		protoClick.DeviceType = click.DeviceType
	}

	return protoClick
}

// ClicksToProto converts a slice of domain Clicks to protobuf Clicks
func ClicksToProto(clicks []domain.Click) []*pb.Click {
	protoClicks := make([]*pb.Click, len(clicks))
	for i, click := range clicks {
		protoClicks[i] = ClickToProto(&click)
	}
	return protoClicks
}
