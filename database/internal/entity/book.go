package entity

import (
	"errors"
	"time"
)

type Book struct {
	ID        string
	Name      string
	Authors   []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrBookNotFound      = errors.New("book not found")
	ErrBookAlreadyExists = errors.New("book already exists")
)
