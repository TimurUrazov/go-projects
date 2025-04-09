-- +goose Up
CREATE INDEX book_idx ON author_book USING HASH (book_id);

-- +goose Down
DROP INDEX book_idx;
