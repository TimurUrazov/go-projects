package library

import (
	"context"

	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/google/uuid"
)

func (l *libraryImpl) RegisterAuthor(ctx context.Context, authorName string) (entity.Author, error) {
	author := entity.Author{
		ID:   uuid.New().String(),
		Name: authorName,
	}
	return l.authorRepository.RegisterAuthor(ctx, author)
}

func (l *libraryImpl) ChangeAuthorInfo(ctx context.Context, id, name string) error {
	return l.authorRepository.ChangeAuthorInfo(ctx, id, name)
}

func (l *libraryImpl) GetAuthorInfo(ctx context.Context, id string) (entity.Author, error) {
	return l.authorRepository.GetAuthorInfo(ctx, id)
}

func (l *libraryImpl) GetAuthorBooks(ctx context.Context, id string) (<-chan entity.Book, <-chan error) {
	return l.authorRepository.GetAuthorBooks(ctx, id)
}
