-- +goose Up
CREATE INDEX book_name_idx ON book USING HASH ("name");

-- +goose Down
DROP INDEX book_name_idx;
