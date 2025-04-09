package controller

import (
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"

	"context"
)

func (i *implementation) RegisterAuthor(ctx context.Context, request *desc.RegisterAuthorRequest) (*desc.RegisterAuthorResponse, error) {
	if err := request.ValidateAll(); err != nil {
		i.logger.Warn("Error validating register author request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorsUseCase.RegisterAuthor(ctx, request.GetName())

	if err != nil {
		i.logger.Debug("Error performing register author use case", zap.Error(err))
		return nil, i.convertErr(err)
	}

	return &desc.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
