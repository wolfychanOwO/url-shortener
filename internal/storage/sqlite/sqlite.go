package sqlite

import (
	"database/sql"
	"fmt"
    
	"main/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
    db *sql.DB
}

func New(storagePath string) (*Storage, error) {
    const op = "storage.sqlite.New"

    db, err := sql.Open("sqlite3", storagePath)
    if err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }

    // shorten http link
    
    stmt, err := db.Prepare(`
        CREATE TABLE IF NOT EXISTS url(
                            id INTEGER PRIMARY KEY,
                            alias TEXT NOT NULL UNIQUE,
                            url TEXT NOT NULL);
        CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
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

func (a *Storage) SaveURL(urlToSave, alias string) (int64, error) {
    const op = "storage.sqlite.SaveURL"

    stmt, err := a.db.Prepare(`
        INSERT INTO url(url, alias) VALUES(?, ?)
    `)
    if err != nil {
        return 0, fmt.Errorf("%s: %w", op, err)
    }

    res, err := stmt.Exec(urlToSave, alias)
    if err != nil {
        if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
            return 0, fmt.Errorf("&s: %w", op, storage.ErrURLExists)
        } 
        return 0, fmt.Errorf("%s: %w", op, err)
    }

    id, err := res.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("%s: falied to get last insert id: %w", op, err)
    }

    return id, err
}

func (s *Storage) GetURL(alias string) (string, error) {
    const op = "storage.sqlite.GetURL"

    stmt, err := s.db.Prepare(`
        SELECT url FROM url WHERE alias = ?
    `)
    if err != nil {
        return "", fmt.Errorf("%s: prepare statement: %w", op, err)
    }

    var resURL string
    err = stmt.QueryRow(alias).Scan(&resURL)

    switch err {
    case sql.ErrNoRows:
        return "", storage.ErrURLNotFound
    case nil:
        return resURL, nil
    default:
        return "", fmt.Errorf("%s: execute statement: %w", op, err)
    }

    /*
    if errors.Is(err, sql.ErrNoRows) {
        return "", storage.ErrURLNotFound
    }
    if err != nil {
        return "", fmt.Errorf("%s: execute statement: %w", op, err)
    }

    return resURL, nil
    */
}

func (s *Storage) DeleteURL(alias string) error {
    const op = "storage.sqlite.DeleteURL"

    stmt, err := s.db.Prepare(`
        DELETE FROM url WHERE alias = ?
    `)
    if err != nil {
        return fmt.Errorf("%s: prepare statement: %w", op, err)
    }

    _, err = stmt.Exec(alias)
    if err != nil {
        return fmt.Errorf("storage.sqlite.DeleteURL: %w", err)
    }
    
    return nil 
}
