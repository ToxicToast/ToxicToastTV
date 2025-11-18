package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto"
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
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	companies, total, err := h.companyUC.ListCompanies(ctx, int(page), int(pageSize), search, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListCompaniesResponse{
		Companies:  mapper.CompaniesToProto(companies),
		Total:      int32(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
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

func (h *CompanyHandler) DeleteCompany(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	err := h.companyUC.DeleteCompany(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			return nil, status.Error(codes.NotFound, "company not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Company deleted successfully",
	}, nil
}
