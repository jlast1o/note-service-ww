package store

import (
	"context"
	"errors"
	"fmt"
	"service/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteStorer interface {
	Create(ctx context.Context, title, content string) (model.Note, error)
	GetAll(ctx context.Context) ([]model.Note, error)
	GetByID(ctx context.Context, id int) (model.Note, error)
}

var ErrNotFound = errors.New("заметка не найдена")

type NoteStore struct {
	pool *pgxpool.Pool
}

func NewNoteStore(pool *pgxpool.Pool) *NoteStore {
	return &NoteStore{pool: pool}
}

func (s *NoteStore) Create(ctx context.Context, title, content string) (model.Note, error) {
	var note model.Note

	query := `
		INSERT INTO notes (title, content)
		VALUES ($1, $2)
		RETURNING id, title, content, created_at
	`

	row := s.pool.QueryRow(ctx, query, title, content) // выполняем запрос

	// Сканируем результат в поля структуры note
	err := row.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
	if err != nil {
		// Обёртываем ошибку, чтобы было понятно, где она возникла
		return model.Note{}, fmt.Errorf("создание заметки: %w", err)
	}

	return note, nil // успешно создана
}

func (s *NoteStore) GetAll(ctx context.Context) ([]model.Note, error) {
	query := `SELECT id, title, content, created_at FROM notes ORDER BY id`

	rows, err := s.pool.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("Не удалось найти записи: %w", err)
	}

	defer rows.Close()

	var notes []model.Note

	for rows.Next() {
		var note model.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt); err != nil {
			return nil, fmt.Errorf("Сканирование строки: %w", err)
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Ошибка интеграции строки: %w", err)
	}

	return notes, nil
}

func (s *NoteStore) GetByID(ctx context.Context, id int) (model.Note, error) {
	query := `SELECT id, title, content, created_at FROM notes where id = $1`

	row := s.pool.QueryRow(ctx, query, id)

	var note model.Note
	err := row.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)

	if err != nil {
		return model.Note{}, fmt.Errorf("поиск заметки: %w", err)
	}

	return note, nil
}
