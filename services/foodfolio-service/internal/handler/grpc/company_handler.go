package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/foodfolio-service/api/proto"
	"toxictoast/services/foodfolio-service/internal/command"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/query"
)

type CompanyHandler struct {
	pb.UnimplementedCompanyServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewCompanyHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *CompanyHandler {
	return &CompanyHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *CompanyHandler) CreateCompany(ctx context.Context, req *pb.CreateCompanyRequest) (*pb.CreateCompanyResponse, error) {
	cmd := &command.CreateCompanyCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCompanyResponse{
		Company: &pb.Company{
			Name: req.Name,
		},
	}, nil
}

func (h *CompanyHandler) GetCompany(ctx context.Context, req *pb.IdRequest) (*pb.GetCompanyResponse, error) {
	qry := &query.GetCompanyByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "company not found")
	}

	company := result.(*domain.Company)

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

	qry := &query.ListCompaniesQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           int(page),
		PageSize:       int(pageSize),
		Search:         search,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListCompaniesResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListCompaniesResponse{
		Companies:  mapper.CompaniesToProto(listResult.Companies),
		Total:      int32(listResult.Total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *CompanyHandler) UpdateCompany(ctx context.Context, req *pb.UpdateCompanyRequest) (*pb.UpdateCompanyResponse, error) {
	cmd := &command.UpdateCompanyCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated company
	qry := &query.GetCompanyByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "company not found")
	}

	company := result.(*domain.Company)

	return &pb.UpdateCompanyResponse{
		Company: mapper.CompanyToProto(company),
	}, nil
}

func (h *CompanyHandler) DeleteCompany(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteCompanyCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Company deleted successfully",
	}, nil
}
