package library

import (
	"context"

	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/TimurUrazov/go-projects/database/internal/usecase/repository"
	"go.uber.org/zap"
)

type AuthorUseCase interface {
	RegisterAuthor(ctx context.Context, authorName string) (entity.Author, error)
	ChangeAuthorInfo(ctx context.Context, id, name string) error
	GetAuthorInfo(ctx context.Context, id string) (entity.Author, error)
	GetAuthorBooks(ctx context.Context, id string) (<-chan entity.Book, <-chan error)
}

type BooksUseCase interface {
	AddBook(ctx context.Context, name string, authorIDs []string) (entity.Book, error)
	UpdateBook(ctx context.Context, id, name string, authorIDs []string) error
	GetBookInfo(ctx context.Context, bookID string) (entity.Book, error)
}

var _ AuthorUseCase = (*libraryImpl)(nil)
var _ BooksUseCase = (*libraryImpl)(nil)

type libraryImpl struct {
	logger           *zap.Logger
	authorRepository repository.AuthorRepository
	booksRepository  repository.BooksRepository
}

func New(
	logger *zap.Logger,
	authorRepository repository.AuthorRepository,
	booksRepository repository.BooksRepository,
) *libraryImpl {
	return &libraryImpl{
		logger:           logger,
		authorRepository: authorRepository,
		booksRepository:  booksRepository,
	}
}
