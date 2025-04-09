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

func Test_implementation_GetAuthorInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		request    *desc.GetAuthorInfoRequest
		setupMocks func(authorUseCase *library.MockAuthorUseCase)
		wantError  bool
		errorCode  codes.Code
	}{
		{
			name: "Successful retrieval of author's info",
			request: &desc.GetAuthorInfoRequest{
				Id: uuid.New().String(),
			},
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.EXPECT().
					GetAuthorInfo(gomock.Any(), gomock.Any()).
					Return(entity.Author{}, nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Author not found",
			request: &desc.GetAuthorInfoRequest{
				Id: uuid.New().String(),
			},
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.EXPECT().
					GetAuthorInfo(gomock.Any(), gomock.Any()).
					Return(entity.Author{}, entity.ErrAuthorNotFound)
			},
			wantError: true,
			errorCode: codes.NotFound,
		},
		{
			name: "Invalid uuid",
			request: &desc.GetAuthorInfoRequest{
				Id: "1",
			},
			setupMocks: nil,
			wantError:  true,
			errorCode:  codes.InvalidArgument,
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
			_, err := impl.GetAuthorInfo(ctx, tt.request)

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
