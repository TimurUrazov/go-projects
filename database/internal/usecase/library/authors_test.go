package library

import (
	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/TimurUrazov/go-projects/database/internal/usecase/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"context"
	"errors"
	"testing"
)

func Test_libraryImpl_RegisterAuthor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		authorName string
		setupMocks func(authorRepository *repository.MockAuthorRepository)
		wantErr    bool
	}{
		{
			name:       "Successfully register author",
			authorName: "Alexander Pushkin",
			setupMocks: func(authorRepository *repository.MockAuthorRepository) {
				authorRepository.EXPECT().
					RegisterAuthor(gomock.Any(), gomock.Any()).
					Return(entity.Author{}, nil)
			},
			wantErr: false,
		},
		{
			name:       "Error while register author",
			authorName: "Zachem vsem znat",
			setupMocks: func(authorRepository *repository.MockAuthorRepository) {
				authorRepository.EXPECT().
					RegisterAuthor(gomock.Any(), gomock.Any()).
					Return(entity.Author{}, errors.New("some repo error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			t.Cleanup(func() {
				ctrl.Finish()
			})

			authorRepository := repository.NewMockAuthorRepository(ctrl)
			booksRepository := repository.NewMockBooksRepository(ctrl)
			logger := zap.NewNop()

			impl := New(logger, authorRepository, booksRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(authorRepository)
			}

			ctx := context.Background()
			_, err := impl.RegisterAuthor(ctx, tt.authorName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_libraryImpl_ChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		authorID   string
		authorName string
		setupMocks func(authorRepository *repository.MockAuthorRepository)
		wantErr    bool
	}{
		{
			name:       "Successfully change author info",
			authorID:   uuid.New().String(),
			authorName: "Alexander Pushkin",
			setupMocks: func(authorRepository *repository.MockAuthorRepository) {
				authorRepository.EXPECT().
					ChangeAuthorInfo(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "Author not found",
			authorID:   uuid.New().String(),
			authorName: "Gleb Copyrkin",
			setupMocks: func(authorRepository *repository.MockAuthorRepository) {
				authorRepository.EXPECT().
					ChangeAuthorInfo(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(entity.ErrAuthorNotFound)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			t.Cleanup(func() {
				ctrl.Finish()
			})

			authorRepository := repository.NewMockAuthorRepository(ctrl)
			booksRepository := repository.NewMockBooksRepository(ctrl)
			logger := zap.NewNop()

			impl := New(logger, authorRepository, booksRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(authorRepository)
			}

			ctx := context.Background()
			err := impl.ChangeAuthorInfo(ctx, tt.authorID, tt.authorName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_libraryImpl_GetAuthorInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		authorID   string
		setupMocks func(authorRepository *repository.MockAuthorRepository)
		wantErr    bool
	}{
		{
			name:     "Successfully get author info",
			authorID: uuid.New().String(),
			setupMocks: func(authorRepository *repository.MockAuthorRepository) {
				authorRepository.EXPECT().
					GetAuthorInfo(gomock.Any(), gomock.Any()).
					Return(entity.Author{}, nil)
			},
			wantErr: false,
		},
		{
			name:     "Author not found",
			authorID: uuid.New().String(),
			setupMocks: func(authorRepository *repository.MockAuthorRepository) {
				authorRepository.EXPECT().
					GetAuthorInfo(gomock.Any(), gomock.Any()).
					Return(entity.Author{}, entity.ErrAuthorNotFound)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			t.Cleanup(func() {
				ctrl.Finish()
			})

			authorRepository := repository.NewMockAuthorRepository(ctrl)
			booksRepository := repository.NewMockBooksRepository(ctrl)
			logger := zap.NewNop()

			impl := New(logger, authorRepository, booksRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(authorRepository)
			}

			ctx := context.Background()
			_, err := impl.GetAuthorInfo(ctx, tt.authorID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_libraryImpl_GetAuthorBooks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		authorID   string
		setupMocks func(authorRepository *repository.MockAuthorRepository)
		wantErr    bool
	}{
		{
			name:     "Successfully get author books",
			authorID: uuid.New().String(),
			setupMocks: func(authorRepository *repository.MockAuthorRepository) {
				authorRepository.EXPECT().
					GetAuthorBooks(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, id string) (<-chan entity.Book, <-chan error) {
						ch := make(chan entity.Book)
						errChan := make(chan error, 1)
						close(errChan)
						go func() {
							defer close(ch)
							ch <- entity.Book{
								Name: "Some Book",
							}
						}()
						return ch, errChan
					})
			},
			wantErr: false,
		},
		{
			name:     "Author not found",
			authorID: uuid.New().String(),
			setupMocks: func(authorRepository *repository.MockAuthorRepository) {
				ch := make(chan entity.Book)
				errChan := make(chan error, 1)
				go func() {
					defer close(ch)
					defer close(errChan)
					errChan <- entity.ErrAuthorNotFound
				}()
				authorRepository.EXPECT().
					GetAuthorBooks(gomock.Any(), gomock.Any()).
					Return(ch, errChan)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			t.Cleanup(func() {
				ctrl.Finish()
			})

			authorRepository := repository.NewMockAuthorRepository(ctrl)
			booksRepository := repository.NewMockBooksRepository(ctrl)
			logger := zap.NewNop()

			impl := New(logger, authorRepository, booksRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(authorRepository)
			}

			ctx := context.Background()
			bookCh, errCh := impl.GetAuthorBooks(ctx, tt.authorID)

			err, ok := <-errCh

			if tt.wantErr {
				if !ok {
					t.Errorf("GetAuthorBooks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else {
				book := <-bookCh
				require.Equal(t, "Some Book", book.Name)
				require.NoError(t, err)
			}
		})
	}
}
