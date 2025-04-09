-- +goose Up
CREATE INDEX author_name_idx ON author USING HASH ("name");

-- +goose Down
DROP INDEX author_name_idx;
