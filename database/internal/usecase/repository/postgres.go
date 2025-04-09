package repository

import (
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"

	"github.com/TimurUrazov/go-projects/database/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"

	"context"
	"errors"
	"strings"
)

var _ BooksRepository = (*postgresRepository)(nil)
var _ AuthorRepository = (*postgresRepository)(nil)

type postgresRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresRepository(db *pgxpool.Pool, logger *zap.Logger) *postgresRepository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}

func (p *postgresRepository) AddBook(ctx context.Context, book entity.Book) (entity.Book, error) {
	tx, err := p.db.Begin(ctx)

	if err != nil {
		p.logger.Warn("Error while starting transaction in add book method", zap.Error(err))
		return entity.Book{}, err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrTxClosed) {
				p.logger.Debug("Tx is closed in add book method", zap.Error(err))
			} else {
				p.logger.Warn("Error while closing transaction in add book method", zap.Error(err))
			}
		}
	}(tx, ctx)

	const queryBook = `INSERT INTO book (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(ctx, queryBook, book.Name).Scan(&book.ID, &book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		p.logger.Warn("Error while performing insert book query in add book method", zap.Error(err))
		return entity.Book{}, err
	}

	const query = `INSERT INTO author_book (author_id, book_id) VALUES ($1, $2)`

	for _, authorID := range book.Authors {
		_, er := tx.Exec(ctx, query, authorID, book.ID)

		var pgErr *pgconn.PgError

		if errors.As(er, &pgErr) && pgErr.Code == "23503" {
			p.logger.Debug("Author not found error while performing insert query in 'author_book' table in add book method",
				zap.String("author_id", authorID),
				zap.Error(er))
			return entity.Book{}, entity.ErrAuthorNotFound
		}

		if er != nil {
			p.logger.Warn("Error while performing insert query in 'author_book' table in add book method",
				zap.Error(er))
			return entity.Book{}, er
		}
	}

	if err = tx.Commit(ctx); err != nil {
		p.logger.Warn("Error while commiting transaction in add book method")
		return entity.Book{}, err
	}

	return book, nil
}

func (p *postgresRepository) GetBookInfo(ctx context.Context, bookID string) (entity.Book, error) {
	tx, err := p.db.Begin(ctx)

	if err != nil {
		p.logger.Warn("Error while starting transaction in get book info method", zap.Error(err))
		return entity.Book{}, err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrTxClosed) {
				p.logger.Debug("Tx is closed in get book info method", zap.Error(err))
			} else {
				p.logger.Warn("Error while closing transaction in get book info method", zap.Error(err))
			}
		}
	}(tx, ctx)

	const query = `SELECT id, name, created_at, updated_at FROM book WHERE id = $1`

	book := entity.Book{}

	err = p.db.QueryRow(ctx, query, bookID).Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		p.logger.Debug("Book not found in select query in get book info method",
			zap.String("book_id", bookID))
		return entity.Book{}, entity.ErrBookNotFound
	}

	if err != nil {
		p.logger.Warn("Error while performing select query to table 'book' in get book info method",
			zap.Error(err))
		return entity.Book{}, err
	}

	const bookAuthorsQuery = `SELECT author_id FROM author_book WHERE book_id = $1`

	rows, err := p.db.Query(ctx, bookAuthorsQuery, bookID)

	if err != nil {
		p.logger.Warn("Error while retrieving authors of book in get book info method",
			zap.String("book_id", bookID), zap.Error(err))
		return entity.Book{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var authorID string

		if err := rows.Scan(&authorID); err != nil {
			p.logger.Warn("Error while scanning author of book in get book info method",
				zap.String("book_id", bookID), zap.String("author_id", authorID), zap.Error(err))
			return entity.Book{}, err
		}

		book.Authors = append(book.Authors, authorID)
	}

	return book, nil
}

func (p *postgresRepository) UpdateBook(ctx context.Context, id, name string, authorIDs []string) error {
	tx, err := p.db.Begin(ctx)

	if err != nil {
		p.logger.Warn("Error while starting transaction in update book method", zap.Error(err))
		return err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrTxClosed) {
				p.logger.Debug("Tx is closed in update book method", zap.Error(err))
			} else {
				p.logger.Warn("Error while closing transaction in update book method", zap.Error(err))
			}
		}
	}(tx, ctx)

	const query = `UPDATE book SET name = $1 WHERE id = $2 RETURNING id`

	var res string

	err = tx.QueryRow(ctx, query, name, id).Scan(&res)

	if errors.Is(err, pgx.ErrNoRows) {
		p.logger.Debug("Book not found in update book method while updating table 'book'",
			zap.String("book_id", id))
		return entity.ErrBookNotFound
	}

	if err != nil {
		p.logger.Warn("Error while performing update book query in update book method",
			zap.String("book_id", id))
		return err
	}

	const queryDeleteBookAuthors = `DELETE FROM author_book WHERE book_id = $1`

	_, err = tx.Exec(ctx, queryDeleteBookAuthors, id)

	if err != nil {
		p.logger.Warn("Error while performing delete book authors query in update book method",
			zap.String("book_id", id))
		return err
	}

	const queryInsertAuthor = `INSERT INTO author_book (book_id, author_id) VALUES ($1, $2)`

	for _, authorID := range authorIDs {
		_, err = tx.Exec(ctx, queryInsertAuthor, id, authorID)

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			p.logger.Debug("Author not found error while inserting author in 'author_book' table in update book method",
				zap.String("author_id", authorID), zap.String("book_id", id))
			return entity.ErrAuthorNotFound
		}

		if err != nil {
			p.logger.Warn("Error while performing insert author in 'author_book' table query in update book method",
				zap.String("author_id", authorID), zap.String("book_id", id), zap.Error(err))
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		p.logger.Warn("Error while commiting transaction in update book method", zap.Error(err))
		return err
	}

	return nil
}

func (p *postgresRepository) ChangeAuthorInfo(ctx context.Context, id, name string) error {
	tx, err := p.db.Begin(ctx)

	if err != nil {
		p.logger.Warn("Error while starting transaction in change author info method", zap.Error(err))
		return err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrTxClosed) {
				p.logger.Debug("Tx is closed in method: change author info", zap.Error(err))
			} else {
				p.logger.Warn("Error while closing transaction in method change author info", zap.Error(err))
			}
		}
	}(tx, ctx)

	const query = `UPDATE author SET name = $1 WHERE id = $2 RETURNING id`

	var res string

	err = tx.QueryRow(ctx, query, name, id).Scan(&res)

	if errors.Is(err, pgx.ErrNoRows) {
		p.logger.Debug("Author not found while updating 'author' table in change author info method",
			zap.String("author_id", id))
		return entity.ErrAuthorNotFound
	}

	if err != nil {
		p.logger.Warn("Error while updating 'author' table in change author info method",
			zap.String("author_id", id), zap.Error(err))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		p.logger.Warn("Error while commiting transaction in change author info method", zap.Error(err))
		return err
	}

	return nil
}

func (p *postgresRepository) RegisterAuthor(ctx context.Context, author entity.Author) (entity.Author, error) {
	tx, err := p.db.Begin(ctx)

	if err != nil {
		p.logger.Warn("Error while starting transaction in register author method", zap.Error(err))
		return entity.Author{}, err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrTxClosed) {
				p.logger.Debug("Tx is closed in register author method", zap.Error(err))
			} else {
				p.logger.Warn("Error while closing transaction in method register author", zap.Error(err))
			}
		}
	}(tx, ctx)

	const query = `INSERT INTO author (name) VALUES ($1) RETURNING id, created_at, updated_at`

	err = tx.QueryRow(ctx, query, author.Name).Scan(&author.ID, &author.CreatedAt, &author.UpdatedAt)

	if err != nil {
		p.logger.Warn("Error while performing insert query in table 'author' in register author method",
			zap.String("author_name", author.Name), zap.Error(err))
		return entity.Author{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		p.logger.Warn("Error while commiting transaction in register author method", zap.Error(err))
		return entity.Author{}, err
	}

	return author, nil
}

func (p *postgresRepository) GetAuthorInfo(ctx context.Context, id string) (entity.Author, error) {
	tx, err := p.db.Begin(ctx)

	if err != nil {
		p.logger.Warn("Error while starting transaction in get author info method", zap.Error(err))
		return entity.Author{}, err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrTxClosed) {
				p.logger.Debug("Tx is closed in get author info method", zap.Error(err))
			} else {
				p.logger.Warn("Error while closing transaction in get author info method", zap.Error(err))
			}
		}
	}(tx, ctx)

	const query = `SELECT id, name, created_at, updated_at FROM author WHERE id = $1`

	author := entity.Author{}

	err = p.db.QueryRow(ctx, query, id).Scan(&author.ID, &author.Name, &author.CreatedAt, &author.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		p.logger.Debug("Author not found error while retrieving author info in get author info method",
			zap.String("id", id))
		return entity.Author{}, entity.ErrAuthorNotFound
	}

	if err != nil {
		p.logger.Warn("Error while retrieving author info in get author info method",
			zap.String("id", id), zap.Error(err))
		return entity.Author{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		p.logger.Warn("Error while commiting transaction in get author info method", zap.Error(err))
		return entity.Author{}, err
	}

	return author, nil
}

func (p *postgresRepository) GetAuthorBooks(ctx context.Context, id string) (<-chan entity.Book, <-chan error) {
	booksChan := make(chan entity.Book)
	errChan := make(chan error, 1)

	go func() {
		tx, err := p.db.Begin(ctx)

		if err != nil {
			p.logger.Warn("Error while starting transaction in get author books method", zap.Error(err))
			errChan <- err
			return
		}

		defer func(tx pgx.Tx, ctx context.Context) {
			err = tx.Rollback(ctx)
			if err != nil {
				if errors.Is(err, pgx.ErrTxClosed) {
					p.logger.Debug("Tx is closed in get author books method", zap.Error(err))
				} else {
					p.logger.Warn("Error while closing transaction in get author books method", zap.Error(err))
				}
			}
		}(tx, ctx)

		defer close(booksChan)
		defer close(errChan)

		const queryDeclareCursor = `
DECLARE curs CURSOR FOR SELECT b1.id, b1.name, b1.created_at, b1.updated_at, string_agg(ab1.author_id::text, '\n') FROM 
(SELECT b.id AS id, b.name AS name, b.created_at AS created_at, b.updated_at AS updated_at FROM
book b JOIN author_book a ON b.id = a.book_id WHERE a.author_id = $1) b1 JOIN author_book ab1 ON ab1.book_id = b1.id
GROUP BY b1.id, b1.name, b1.created_at, b1.updated_at
`
		_, err = tx.Exec(ctx, queryDeclareCursor, id)

		if err != nil {
			p.logger.Warn("Error while declaring cursor in get author books method",
				zap.String("author_id", id), zap.Error(err))
			errChan <- err
			return
		}

		rows, err := tx.Query(ctx, "FETCH FORWARD ALL FROM curs")

		if err != nil {
			p.logger.Warn("Error while fetching cursor in get author books method",
				zap.String("author_id", id), zap.Error(err))
			errChan <- err
			return
		}

		defer rows.Close()

		for rows.Next() {
			book := entity.Book{}

			var authors string

			if err := rows.Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt, &authors); err != nil {
				p.logger.Warn("Error while scanning row cursor pointing on in get author books method",
					zap.String("author_id", id), zap.Error(err))
				errChan <- err
				return
			}

			book.Authors = strings.Split(authors, "\\n")

			booksChan <- book
		}

		if err := tx.Commit(ctx); err != nil {
			p.logger.Warn("Error while commiting transaction in get author books method", zap.Error(err))
			errChan <- err
			return
		}
	}()

	return booksChan, errChan
}
