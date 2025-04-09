package controller

import (
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"

	"context"
)

func (i *implementation) UpdateBook(ctx context.Context, req *desc.UpdateBookRequest) (*desc.UpdateBookResponse, error) {
	if err := req.ValidateAll(); err != nil {
		i.logger.Warn("Error validating update book request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.booksUseCase.UpdateBook(ctx, req.GetId(), req.GetName(), req.GetAuthorIds())

	if err != nil {
		i.logger.Debug("Error performing update book use case", zap.Error(err))
		return nil, i.convertErr(err)
	}

	return &desc.UpdateBookResponse{}, nil
}
