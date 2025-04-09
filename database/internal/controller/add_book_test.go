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
	"slices"
	"testing"
)

func Test_implementation_AddBook(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		request    *desc.AddBookRequest
		setupMocks func(booksUseCase *library.MockBooksUseCase)
		wantError  bool
		errorCode  codes.Code
	}{
		{
			name: "Successful book addition!",
			request: &desc.AddBookRequest{
				Name:      "Manifesto of the Communist Party!!!",
				AuthorIds: slices.Repeat([]string{uuid.New().String()}, 2),
			},
			setupMocks: func(booksUseCase *library.MockBooksUseCase) {
				booksUseCase.EXPECT().
					AddBook(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(entity.Book{}, nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Author is not mentioned.",
			request: &desc.AddBookRequest{
				Name:      "Pravilo",
				AuthorIds: []string{},
			},
			setupMocks: func(booksUseCase *library.MockBooksUseCase) {
				booksUseCase.EXPECT().
					AddBook(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(entity.Book{}, nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Many authors",
			request: &desc.AddBookRequest{
				Name:      "American Psycho",
				AuthorIds: slices.Repeat([]string{uuid.New().String()}, 20),
			},
			setupMocks: func(booksUseCase *library.MockBooksUseCase) {
				booksUseCase.EXPECT().
					AddBook(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(entity.Book{}, nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Author does not exist",
			request: &desc.AddBookRequest{
				Name:      "Sviatoslav Korneev",
				AuthorIds: []string{},
			},
			setupMocks: func(booksUseCase *library.MockBooksUseCase) {
				booksUseCase.EXPECT().
					AddBook(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(entity.Book{}, entity.ErrAuthorNotFound)
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
			_, err := impl.AddBook(ctx, tt.request)

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
