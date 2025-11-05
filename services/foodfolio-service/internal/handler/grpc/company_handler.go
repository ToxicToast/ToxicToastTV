package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

type CompanyHandler struct {
	pb.UnimplementedCompanyServiceServer
	companyUC usecase.CompanyUseCase
}

func NewCompanyHandler(companyUC usecase.CompanyUseCase) *CompanyHandler {
	return &CompanyHandler{
		companyUC: companyUC,
	}
}

func (h *CompanyHandler) CreateCompany(ctx context.Context, req *pb.CreateCompanyRequest) (*pb.CreateCompanyResponse, error) {
	company, err := h.companyUC.CreateCompany(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCompanyResponse{
		Company: mapper.CompanyToProto(company),
	}, nil
}

func (h *CompanyHandler) GetCompany(ctx context.Context, req *pb.IdRequest) (*pb.GetCompanyResponse, error) {
	company, err := h.companyUC.GetCompanyByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			return nil, status.Error(codes.NotFound, "company not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetCompanyResponse{
		Company: mapper.CompanyToProto(company),
	}, nil
}

func (h *CompanyHandler) ListCompanies(ctx context.Context, req *pb.ListCompaniesRequest) (*pb.ListCompaniesResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	companies, total, err := h.companyUC.ListCompanies(ctx, page, pageSize, search, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListCompaniesResponse{
		Companies:  mapper.CompaniesToProto(companies),
		Pagination: mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *CompanyHandler) UpdateCompany(ctx context.Context, req *pb.UpdateCompanyRequest) (*pb.UpdateCompanyResponse, error) {
	company, err := h.companyUC.UpdateCompany(ctx, req.Id, req.Name)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			return nil, status.Error(codes.NotFound, "company not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateCompanyResponse{
		Company: mapper.CompanyToProto(company),
	}, nil
}

func (h *CompanyHandler) DeleteCompany(ctx context.Context, req *pb.IdRequest) (*pb.SuccessResponse, error) {
	err := h.companyUC.DeleteCompany(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			return nil, status.Error(codes.NotFound, "company not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SuccessResponse{
		Success: true,
		Message: "Company deleted successfully",
	}, nil
}
