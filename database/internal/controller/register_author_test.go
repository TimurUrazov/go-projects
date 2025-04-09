package controller

import (
	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"
	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/TimurUrazov/go-projects/database/internal/usecase/library"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"
	"errors"
	"testing"
)

func Test_implementation_RegisterAuthor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		request    *desc.RegisterAuthorRequest
		setupMocks func(authorUseCase *library.MockAuthorUseCase)
		wantError  bool
		errorCode  codes.Code
	}{
		{
			name: "Author with valid name",
			request: &desc.RegisterAuthorRequest{
				Name: "Georgy Korneev",
			},
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.EXPECT().
					RegisterAuthor(gomock.Any(), gomock.Any()).
					Return(entity.Author{}, nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Author with invalid name",
			request: &desc.RegisterAuthorRequest{
				Name: "Georgу Korneev",
			},
			setupMocks: nil,
			wantError:  true,
			errorCode:  codes.InvalidArgument,
		},
		{
			name: "Some use case error",
			request: &desc.RegisterAuthorRequest{
				Name: "Steve Apple 2",
			},
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.EXPECT().
					RegisterAuthor(gomock.Any(), gomock.Any()).
					Return(entity.Author{}, errors.New("some use case error"))
			},
			wantError: true,
			errorCode: codes.Internal,
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
				tt.setupMocks(authorUseCase)
			}

			ctx := context.Background()
			_, err := impl.RegisterAuthor(ctx, tt.request)

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
