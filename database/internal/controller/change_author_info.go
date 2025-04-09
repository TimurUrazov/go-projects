package controller

import (
	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"
)

func (i *implementation) ChangeAuthorInfo(ctx context.Context, request *desc.ChangeAuthorInfoRequest) (*desc.ChangeAuthorInfoResponse, error) {
	if err := request.ValidateAll(); err != nil {
		i.logger.Warn("Error validating change author info request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.authorsUseCase.ChangeAuthorInfo(ctx, request.GetId(), request.GetName())

	if err != nil {
		i.logger.Debug("Error performing change author info use case", zap.Error(err))
		return nil, i.convertErr(err)
	}

	return &desc.ChangeAuthorInfoResponse{}, nil
}
