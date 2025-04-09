package controller

import (
	"go.uber.org/zap"

	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"context"
)

func (i *implementation) GetBookInfo(ctx context.Context, request *desc.GetBookInfoRequest) (*desc.GetBookInfoResponse, error) {
	if err := request.ValidateAll(); err != nil {
		i.logger.Warn("Error validating get book info request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.GetBookInfo(ctx, request.GetId())

	if err != nil {
		i.logger.Debug("Error performing get book info use case", zap.Error(err))
		return nil, i.convertErr(err)
	}

	return &desc.GetBookInfoResponse{
		Book: &desc.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.Authors,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
