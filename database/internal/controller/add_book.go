package controller

import (
	"go.uber.org/zap"

	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"

	"context"
)

func (i *implementation) AddBook(ctx context.Context, request *desc.AddBookRequest) (*desc.AddBookResponse, error) {
	if err := request.ValidateAll(); err != nil {
		i.logger.Warn("error validating add book request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.AddBook(ctx, request.GetName(), request.GetAuthorIds())

	if err != nil {
		i.logger.Debug("error performing add book use case", zap.Error(err))
		return nil, i.convertErr(err)
	}

	return &desc.AddBookResponse{
		Book: &desc.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.Authors,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
