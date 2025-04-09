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
	"strings"
	"testing"
)

func Test_implementation_ChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		request    *desc.ChangeAuthorInfoRequest
		setupMocks func(authorUseCase *library.MockAuthorUseCase)
		wantError  bool
		errorCode  codes.Code
	}{
		{
			name: "Author with valid uuid",
			request: &desc.ChangeAuthorInfoRequest{
				Id:   uuid.New().String(),
				Name: "Winston Churchill",
			},
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.EXPECT().
					ChangeAuthorInfo(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Author with invalid uuid",
			request: &desc.ChangeAuthorInfoRequest{
				Id:   "1",
				Name: "Winston Churchill",
			},
			setupMocks: nil,
			wantError:  true,
			errorCode:  codes.InvalidArgument,
		},
		{
			name: "Author with long name",
			request: &desc.ChangeAuthorInfoRequest{
				Id:   uuid.New().String(),
				Name: strings.Repeat("Jean-Paul Sartre", 512),
			},
			setupMocks: nil,
			wantError:  true,
			errorCode:  codes.InvalidArgument,
		},
		{
			name: "Author is noname",
			request: &desc.ChangeAuthorInfoRequest{
				Id:   uuid.New().String(),
				Name: "",
			},
			setupMocks: nil,
			wantError:  true,
			errorCode:  codes.InvalidArgument,
		},
		{
			name: "Author with valid name 512 chars long",
			request: &desc.ChangeAuthorInfoRequest{
				Id:   uuid.New().String(),
				Name: strings.Repeat("W", 512),
			},
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.EXPECT().
					ChangeAuthorInfo(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Author with valid name 512 chars long",
			request: &desc.ChangeAuthorInfoRequest{
				Id:   uuid.New().String(),
				Name: strings.Repeat("W", 512),
			},
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.EXPECT().
					ChangeAuthorInfo(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantError: false,
			errorCode: codes.OK,
		},
		{
			name: "Author with valid name, but not found",
			request: &desc.ChangeAuthorInfoRequest{
				Id:   uuid.New().String(),
				Name: strings.Repeat("W", 250),
			},
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.EXPECT().
					ChangeAuthorInfo(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(entity.ErrAuthorNotFound)
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
				tt.setupMocks(authorUseCase)
			}

			ctx := context.Background()
			_, err := impl.ChangeAuthorInfo(ctx, tt.request)

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
