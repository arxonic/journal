package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/arxonic/journal/internal/domain/models"
	"github.com/arxonic/journal/internal/storage"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) UserRole(email string) (models.Key, error) {
	const fn = "storage.sqlite.UserRole"

	stmt, err := s.db.Prepare("SELECT id, email, role FROM users WHERE email = ?")
	if err != nil {
		return models.Key{}, err
	}

	var key models.Key
	err = stmt.QueryRow(email).Scan(&key.ID, &key.Email, &key.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Key{}, storage.ErrUserNotFound
		}
		return models.Key{}, fmt.Errorf("%s:%w", fn, err)
	}

	return key, nil
}
