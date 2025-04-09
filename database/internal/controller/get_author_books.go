package controller

import (
	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *implementation) GetAuthorBooks(request *desc.GetAuthorBooksRequest, stream desc.Library_GetAuthorBooksServer) error {
	if err := request.ValidateAll(); err != nil {
		i.logger.Warn("error validating get author books request", zap.Error(err))
		return status.Error(codes.InvalidArgument, err.Error())
	}

	booksCh, errCh := i.authorsUseCase.GetAuthorBooks(stream.Context(), request.GetAuthorId())

	for book := range booksCh {
		if err := stream.Send(&desc.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.Authors,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		}); err != nil {
			if st, ok := status.FromError(err); ok {
				i.logger.Debug("Error while performing server streaming", zap.Error(err))
				return status.Error(st.Code(), st.Message())
			}
			i.logger.Warn("Internal error while performing server streaming", zap.Error(err))
			return status.Error(codes.Internal, err.Error())
		}
	}

	if err := <-errCh; err != nil {
		i.logger.Debug("Error performing get author books use case", zap.Error(err))
		return i.convertErr(err)
	}

	return nil
}
