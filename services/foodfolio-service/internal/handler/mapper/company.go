package mapper

import (
	"toxictoast/services/foodfolio-service/internal/domain"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

// CompanyToProto converts domain Company to protobuf
func CompanyToProto(company *domain.Company) *pb.Company {
	if company == nil {
		return nil
	}

	return &pb.Company{
		Id:         company.ID,
		Name:       company.Name,
		Slug:       company.Slug,
		Timestamps: ToTimestamps(company.CreatedAt, company.UpdatedAt, nil),
		ItemCount:  0, // TODO: Calculate if needed
	}
}

// CompaniesToProto converts slice of Companies
func CompaniesToProto(companies []*domain.Company) []*pb.Company {
	result := make([]*pb.Company, len(companies))
	for i, company := range companies {
		result[i] = CompanyToProto(company)
	}
	return result
}
