package controller

import (
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"

	"context"
)

func (i *implementation) GetAuthorInfo(ctx context.Context, req *desc.GetAuthorInfoRequest) (*desc.GetAuthorInfoResponse, error) {
	if err := req.ValidateAll(); err != nil {
		i.logger.Warn("Error validating get author info request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorsUseCase.GetAuthorInfo(ctx, req.GetId())

	if err != nil {
		i.logger.Debug("Error performing change author info use case", zap.Error(err))
		return nil, i.convertErr(err)
	}

	return &desc.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}
