package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"

	"url-minimizer/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	create table if not exists url(
		id integer primary key,
		alias text not null unique,
		url text not null);
	create index if not exists idx_alias on url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Save(url string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveUrl"

	stmt, err := s.db.Prepare(`
	insert into url(url, alias)
	values(?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	res, err := stmt.Exec(url, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) Get(alias string) (string, error) {
	const op = "storage.sqlite.GetUrl"

	stmt, err := s.db.Prepare(`
	select url from url where alias = ?
	`)
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resUrl string
	err = stmt.QueryRow(alias).Scan(&resUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resUrl, nil
}

func (s *Storage) Delete(url string) (int64, error) {
	const op = "storage.sqlite.DeleteUrl"

	stmt, err := s.db.Prepare(`
	delete from url where url = ?
	`)
	if err != nil {
		return -1, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	res, err := stmt.Exec(url)
	if err != nil {
		return -1, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	rowsDeleted, err := res.RowsAffected()
	if err != nil {
		return -1, fmt.Errorf("%s: failed to get rows affected by last delete: %w", op, err)
	}

	return rowsDeleted, nil
}

func (s *Storage) Close() error {
	const op = "storage.sqlite.Close"

	errs := make([]string, 0, 1)
	if err := s.db.Close(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s: %s", op, strings.Join(errs, ", "))
	}
	return nil
}
