package store

import (
	"context"
	"errors"
	"fmt"
	"service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteStorer interface {
	Create(ctx context.Context, title, content string) (model.Note, error)
	GetAll(ctx context.Context, limit, offset int) ([]model.Note, int, error)
	GetByID(ctx context.Context, id int) (model.Note, error)
	Update(ctx context.Context, id int, title, content string) (model.Note, error)
	Delete(ctx context.Context, id int) error
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

func (s *NoteStore) GetAll(ctx context.Context, limit, offset int) ([]model.Note, int, error) {
	var total int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) from notes`).Scan(&total)

	if err != nil {
		return nil, 0, fmt.Errorf("подсчет заметок %w", err)
	}

	query := `SELECT id, title, content, created_at FROM notes ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("получаем заметки: %w", err)
	}
	defer rows.Close()

	var notes []model.Note

	for rows.Next() {
		var note model.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("сканирование строки: %w", err)
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка итерации строк: %w", err)
	}

	return notes, total, nil

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

func (s *NoteStore) Update(ctx context.Context, id int, title, content string) (model.Note, error) {
	query := `UPDATE notes SET title = $1, content = $2 where id = $3 RETURNING id, title, content, created_at`

	row := s.pool.QueryRow(ctx, query, id, title, content)
	var note model.Note

	err := row.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Note{}, ErrNotFound
	}

	if err != nil {
		return model.Note{}, fmt.Errorf("Обновление записи: %w", err)
	}

	return note, nil
}

func (s *NoteStore) Delete(ctx context.Context, id int) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM notes where id = $1`, id)

	if err != nil {
		return fmt.Errorf("Ошибка удаления: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
