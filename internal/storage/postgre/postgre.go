package postgre

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"shorter/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	const op = "storage.postgre.New"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s ping: %w", op, err)
	}
	schema := `
    CREATE TABLE IF NOT EXISTS url(
        id SERIAL PRIMARY KEY,
        alias TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
    `
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("%s init schema: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveUrl(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgre.SaveUrl"
	const query = "INSERT INTO url(alias, url) VALUES($1, $2)"
	trx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s, begin transaction: %w", op, err)
	}
	defer func() {
		if err := trx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("Error: %s", err)
		}
	}()
	stmt, err := trx.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.Exec(alias, urlToSave)
	if err != nil {
		if trxErr := trx.Rollback(); trxErr != nil {
			return 0, fmt.Errorf("%s: %w", op, trxErr)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := res.RowsAffected()
	if err != nil {
		if trxErr := trx.Rollback(); trxErr != nil {
			return 0, fmt.Errorf("%s: %w", op, trxErr)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	err = trx.Commit()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetUrl(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = $1")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return resURL, nil
}
