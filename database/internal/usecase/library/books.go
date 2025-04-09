package library

import (
	"context"

	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/google/uuid"
)

func (l *libraryImpl) AddBook(ctx context.Context, name string, authorIDs []string) (entity.Book, error) {
	book := entity.Book{
		ID:      uuid.New().String(),
		Name:    name,
		Authors: authorIDs,
	}
	return l.booksRepository.AddBook(ctx, book)
}

func (l *libraryImpl) UpdateBook(ctx context.Context, id, name string, authorIDs []string) error {
	return l.booksRepository.UpdateBook(ctx, id, name, authorIDs)
}

func (l *libraryImpl) GetBookInfo(ctx context.Context, bookID string) (entity.Book, error) {
	return l.booksRepository.GetBookInfo(ctx, bookID)
}
