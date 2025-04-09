package controller

import (
	desc "github.com/TimurUrazov/go-projects/database/generated/api/library"
	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/TimurUrazov/go-projects/database/internal/usecase/library"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"
	"errors"
	"sort"
	"testing"
)

var ErrStreamError = errors.New("stream error")

type serverStreamingServerImpl[Res any] struct {
	grpc.ServerStream
	ch    chan<- *Res
	limit int
}

func newServerStreamingServer[Res any](ch chan *Res, limit int) *serverStreamingServerImpl[Res] {
	return &serverStreamingServerImpl[Res]{
		ch:    ch,
		limit: limit,
	}
}

func (ss *serverStreamingServerImpl[Res]) Context() context.Context {
	return context.Background()
}

func (ss *serverStreamingServerImpl[Res]) Send(res *Res) error {
	if ss.limit == 0 {
		return ErrStreamError
	}
	ss.limit--
	ss.ch <- res
	if ss.limit == 0 {
		close(ss.ch)
	}
	return nil
}

type errorStreamingServerImpl[Res any] struct {
	grpc.ServerStream
	err error
}

func (ss *errorStreamingServerImpl[Res]) Context() context.Context {
	return context.Background()
}

func newErrorStreamingServer[Res any](err error) *errorStreamingServerImpl[Res] {
	return &errorStreamingServerImpl[Res]{
		err: err,
	}
}

func (ss *errorStreamingServerImpl[Res]) Send(_ *Res) error {
	return ss.err
}

func Test_implementation_GetAuthorBooks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		setupMocks func(authorUseCase *library.MockAuthorUseCase)
		action     func(t *testing.T, impl *implementation)
	}{
		{
			name: "Successful retrieval of author's books",
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				useCaseResults := []entity.Book{
					{Name: "My Universities"},
					{Name: "The Lower Depths"},
				}
				authorUseCase.EXPECT().
					GetAuthorBooks(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, id string) (<-chan entity.Book, <-chan error) {
						ch := make(chan entity.Book)
						errChan := make(chan error, 1)
						go func() {
							defer close(ch)
							defer close(errChan)
							for _, r := range useCaseResults {
								ch <- r
							}
						}()
						return ch, errChan
					})
			},
			action: func(t *testing.T, impl *implementation) {
				t.Helper()
				serviceCh := make(chan *desc.Book)
				go func() {
					err := impl.GetAuthorBooks(&desc.GetAuthorBooksRequest{
						AuthorId: uuid.New().String(),
					}, newServerStreamingServer(serviceCh, 2))
					assert.NoError(t, err)
				}()
				bookNames := make([]string, 0)
				for res := range serviceCh {
					bookNames = append(bookNames, res.GetName())
				}
				sort.Strings(bookNames)
				require.Equal(t, []string{"My Universities", "The Lower Depths"}, bookNames)
			},
		},
		{
			name: "Author not found",
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.
					EXPECT().
					GetAuthorBooks(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, id string) (<-chan entity.Book, <-chan error) {
						ch := make(chan entity.Book)
						errChan := make(chan error, 1)
						errChan <- entity.ErrAuthorNotFound
						defer close(ch)
						defer close(errChan)
						return ch, errChan
					})
			},
			action: func(t *testing.T, impl *implementation) {
				t.Helper()
				err := impl.GetAuthorBooks(&desc.GetAuthorBooksRequest{
					AuthorId: uuid.New().String(),
				}, newServerStreamingServer(make(chan *desc.Book), 0)) // limit is 0, so error arises
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "Author books stream error",
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.
					EXPECT().
					GetAuthorBooks(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, id string) (<-chan entity.Book, <-chan error) {
						ch := make(chan entity.Book)
						errChan := make(chan error, 1)
						go func() {
							defer close(ch)
							defer close(errChan)
							ch <- entity.Book{}
						}()
						return ch, errChan
					})
			},
			action: func(t *testing.T, impl *implementation) {
				t.Helper()
				err := impl.GetAuthorBooks(&desc.GetAuthorBooksRequest{
					AuthorId: uuid.New().String(),
				}, newServerStreamingServer(make(chan *desc.Book), 0))
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
				require.ErrorContains(t, err, ErrStreamError.Error())
			},
		},
		{
			name: "Get author books resource exhausted",
			setupMocks: func(authorUseCase *library.MockAuthorUseCase) {
				authorUseCase.
					EXPECT().
					GetAuthorBooks(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, id string) (<-chan entity.Book, <-chan error) {
						ch := make(chan entity.Book)
						errChan := make(chan error, 1)
						go func() {
							defer close(ch)
							defer close(errChan)
							ch <- entity.Book{}
						}()
						return ch, errChan
					})
			},
			action: func(t *testing.T, impl *implementation) {
				t.Helper()
				err := impl.GetAuthorBooks(&desc.GetAuthorBooksRequest{
					AuthorId: uuid.New().String(),
				}, newErrorStreamingServer[desc.Book](status.Errorf(codes.ResourceExhausted, "grpc: message too large")))
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.ResourceExhausted, st.Code())
			},
		},
		{
			name:       "Get invalid author books request",
			setupMocks: nil,
			action: func(t *testing.T, impl *implementation) {
				t.Helper()
				err := impl.GetAuthorBooks(&desc.GetAuthorBooksRequest{
					AuthorId: "1",
				}, nil)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
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

			tt.action(t, impl)
		})
	}
}
