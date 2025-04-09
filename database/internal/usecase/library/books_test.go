package library

import (
	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/TimurUrazov/go-projects/database/internal/usecase/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"context"
	"testing"
)

func Test_libraryImpl_AddBook(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		bookName   string
		authorIDs  []string
		setupMocks func(booksRepository *repository.MockBooksRepository)
		wantErr    bool
	}{
		{
			name:      "Successful book addition",
			bookName:  "Ahahahaha",
			authorIDs: []string{"Lermontov"},
			setupMocks: func(booksRepository *repository.MockBooksRepository) {
				booksRepository.EXPECT().
					AddBook(gomock.Any(), gomock.Any()).
					Return(entity.Book{}, nil)
			},
			wantErr: false,
		},
		{
			name:      "Author not found",
			bookName:  "He is really dead",
			authorIDs: []string{"You Will Never Know"},
			setupMocks: func(booksRepository *repository.MockBooksRepository) {
				booksRepository.EXPECT().
					AddBook(gomock.Any(), gomock.Any()).
					Return(entity.Book{}, entity.ErrAuthorNotFound)
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
				tt.setupMocks(booksRepository)
			}

			ctx := context.Background()
			_, err := impl.AddBook(ctx, tt.bookName, tt.authorIDs)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_libraryImpl_UpdateBook(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		bookID     string
		bookName   string
		authorIDs  []string
		setupMocks func(booksRepository *repository.MockBooksRepository)
		wantErr    bool
	}{
		{
			name:      "Successful book info retrieval",
			bookID:    uuid.New().String(),
			bookName:  "You are genius!",
			authorIDs: []string{"You Yes Really You"},
			setupMocks: func(booksRepository *repository.MockBooksRepository) {
				booksRepository.EXPECT().
					UpdateBook(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "Book not found",
			bookID:    uuid.New().String(),
			bookName:  "Stalker, go out!",
			authorIDs: []string{"You Know His Thin Voice", "And His Crazy Laugh"},
			setupMocks: func(booksRepository *repository.MockBooksRepository) {
				booksRepository.EXPECT().
					UpdateBook(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(entity.ErrBookNotFound)
			},
			wantErr: true,
		},
		{
			name:      "Author not found",
			bookID:    uuid.New().String(),
			bookName:  "Gleb SVO",
			authorIDs: []string{"What A Pity"},
			setupMocks: func(booksRepository *repository.MockBooksRepository) {
				booksRepository.EXPECT().
					UpdateBook(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
				tt.setupMocks(booksRepository)
			}

			ctx := context.Background()
			err := impl.UpdateBook(ctx, tt.bookID, tt.bookName, tt.authorIDs)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_libraryImpl_GetBookInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		bookID     string
		setupMocks func(booksRepository *repository.MockBooksRepository)
		wantErr    bool
	}{
		{
			name:   "Successful book info retrieval",
			bookID: uuid.New().String(),
			setupMocks: func(booksRepository *repository.MockBooksRepository) {
				booksRepository.EXPECT().
					GetBookInfo(gomock.Any(), gomock.Any()).
					Return(entity.Book{}, nil)
			},
			wantErr: false,
		},
		{
			name:   "Book not found",
			bookID: uuid.New().String(),
			setupMocks: func(booksRepository *repository.MockBooksRepository) {
				booksRepository.EXPECT().
					GetBookInfo(gomock.Any(), gomock.Any()).
					Return(entity.Book{}, entity.ErrBookNotFound)
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
				tt.setupMocks(booksRepository)
			}

			ctx := context.Background()
			_, err := impl.GetBookInfo(ctx, tt.bookID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
