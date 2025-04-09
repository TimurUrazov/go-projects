package controller

import (
	generated "github.com/TimurUrazov/go-projects/database/generated/api/library"
	"github.com/TimurUrazov/go-projects/database/internal/usecase/library"
	"go.uber.org/zap"
)

var _ generated.LibraryServer = (*implementation)(nil)

type implementation struct {
	logger         *zap.Logger
	booksUseCase   library.BooksUseCase
	authorsUseCase library.AuthorUseCase
}

func New(
	logger *zap.Logger,
	booksUseCase library.BooksUseCase,
	authorsUseCase library.AuthorUseCase,
) *implementation {
	return &implementation{
		logger:         logger,
		booksUseCase:   booksUseCase,
		authorsUseCase: authorsUseCase,
	}
}
