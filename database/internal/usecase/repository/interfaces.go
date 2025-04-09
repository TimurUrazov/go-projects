package repository

import (
	"context"

	"github.com/TimurUrazov/go-projects/database/internal/entity"
)

type (
	AuthorRepository interface {
		RegisterAuthor(ctx context.Context, name entity.Author) (entity.Author, error)
		ChangeAuthorInfo(ctx context.Context, id, name string) error
		GetAuthorInfo(ctx context.Context, id string) (entity.Author, error)
		GetAuthorBooks(ctx context.Context, id string) (<-chan entity.Book, <-chan error)
	}

	BooksRepository interface {
		AddBook(ctx context.Context, book entity.Book) (entity.Book, error)
		UpdateBook(ctx context.Context, id, name string, authorIDs []string) error
		GetBookInfo(ctx context.Context, bookID string) (entity.Book, error)
	}
)
