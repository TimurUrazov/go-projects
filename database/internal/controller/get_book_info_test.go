package controller

import (
	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"
	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/TimurUrazov/go-projects/database/internal/usecase/library"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"
	"testing"
)

func Test_implementation_GetBookInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		request    *desc.GetBookInfoRequest
		setupMocks func(booksUseCase *library.MockBooksUseCase)
		wantError  bool
		errorCode  codes.Code
	}{
		{
			name: "Successful book info retrieval",
			request: &desc.GetBookInfoRequest{
				Id: uuid.New().String(),
			},
			setupMocks: func(booksUseCase *library.MockBooksUseCase) {
				booksUseCase.EXPECT().
					GetBookInfo(gomock.Any(), gomock.Any()).
					Return(entity.Book{}, nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Invalid uuid",
			request: &desc.GetBookInfoRequest{
				Id: "1",
			},
			setupMocks: nil,
			wantError:  true,
			errorCode:  codes.InvalidArgument,
		},
		{
			name: "Book not found",
			request: &desc.GetBookInfoRequest{
				Id: uuid.New().String(),
			},
			setupMocks: func(booksUseCase *library.MockBooksUseCase) {
				booksUseCase.EXPECT().
					GetBookInfo(gomock.Any(), gomock.Any()).
					Return(entity.Book{}, entity.ErrBookNotFound)
			},
			wantError: true,
			errorCode: codes.NotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			t.Cleanup(func() {
				ctrl.Finish()
			})

			authorUseCase := library.NewMockAuthorUseCase(ctrl)
			bookUseCase := library.NewMockBooksUseCase(ctrl)
			logger := zap.NewNop()

			impl := New(logger, bookUseCase, authorUseCase)

			if tt.setupMocks != nil {
				tt.setupMocks(bookUseCase)
			}

			ctx := context.Background()
			_, err := impl.GetBookInfo(ctx, tt.request)

			st, ok := status.FromError(err)

			if tt.wantError {
				require.True(t, ok)
				require.Equal(t, tt.errorCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
