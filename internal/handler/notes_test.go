package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"service/internal/model"
	"service/internal/store"
	"testing"

	"github.com/go-chi/chi"
)

type mockNoteStore struct {
	createFn  func(ctx context.Context, title, content string) (model.Note, error)
	getAllFn  func(ctx context.Context) ([]model.Note, error)
	getByIDFn func(ctx context.Context, id int) (model.Note, error)
}

func (m *mockNoteStore) Create(ctx context.Context, title, content string) (model.Note, error) {
	return m.createFn(ctx, title, content)
}

func (m *mockNoteStore) GetAll(ctx context.Context) ([]model.Note, error) {
	return m.getAllFn(ctx)
}

func (m *mockNoteStore) GetByID(ctx context.Context, id int) (model.Note, error) {
	return m.getByIDFn(ctx, id)
}

func TestCreateNote_Success(t *testing.T) {
	// Готовим мок, который возвращает успешную заметку
	mockStore := &mockNoteStore{
		createFn: func(ctx context.Context, title, content string) (model.Note, error) {
			return model.Note{ID: 1, Title: title, Content: content}, nil // имитируем создание с ID=1
		},
	}
	handler := NewNoteHandler(mockStore) // создаём обработчик с моком

	// Готовим тело запроса (правильный JSON)
	body := `{"title":"Test Title","content":"Test Content"}`
	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(body)) // создаём запрос
	req.Header.Set("Content-Type", "application/json")                                 // указываем контент-тип
	w := httptest.NewRecorder()                                                        // фиктивный ResponseWriter

	handler.Create(w, req) // вызываем обработчик

	resp := w.Result() // получаем *http.Response
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated { // ожидаем 201
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var note model.Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil { // декодируем ответ
		t.Fatalf("failed to decode response: %v", err)
	}
	if note.ID != 1 || note.Title != "Test Title" { // проверяем поля
		t.Errorf("unexpected note: %+v", note)
	}
}

func TestCreateNote_InvalidJSON(t *testing.T) {
	mockStore := &mockNoteStore{} // мок не будет вызван
	handler := NewNoteHandler(mockStore)

	body := `invalid`
	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest { // ожидаем 400
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateNote_EmptyTitle(t *testing.T) {
	mockStore := &mockNoteStore{} // не должен создавать
	handler := NewNoteHandler(mockStore)

	body := `{"title":"","content":"whatever"}`
	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateNote_StoreError(t *testing.T) {
	// Мок, который возвращает ошибку БД
	mockStore := &mockNoteStore{
		createFn: func(ctx context.Context, title, content string) (model.Note, error) {
			return model.Note{}, errors.New("db connection lost")
		},
	}
	handler := NewNoteHandler(mockStore)

	body := `{"title":"valid"}`
	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestGetAll_Success(t *testing.T) {
	mockStore := &mockNoteStore{
		getAllFn: func(ctx context.Context) ([]model.Note, error) {
			return []model.Note{
				{ID: 1, Title: "first"},
				{ID: 2, Title: "second"},
			}, nil
		},
	}
	handler := NewNoteHandler(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()
	handler.GetAll(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var notes []model.Note
	json.NewDecoder(resp.Body).Decode(&notes)
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
}

func TestGetAll_Empty(t *testing.T) {
	mockStore := &mockNoteStore{
		getAllFn: func(ctx context.Context) ([]model.Note, error) {
			return nil, nil // вернёт nil, а обработчик превратит в []
		},
	}
	handler := NewNoteHandler(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()
	handler.GetAll(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var notes []model.Note
	json.NewDecoder(resp.Body).Decode(&notes)
	if notes == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

func TestGetByID_Success(t *testing.T) {
	mockStore := &mockNoteStore{
		getByIDFn: func(ctx context.Context, id int) (model.Note, error) {
			return model.Note{ID: id, Title: "found"}, nil
		},
	}
	handler := NewNoteHandler(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/notes/123", nil)
	// chi требует контекст с параметром, поэтому симулируем через chi.RouteContext
	req = req.WithContext(setChiParam(req.Context(), "id", "123"))
	w := httptest.NewRecorder()
	handler.GetByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var note model.Note
	json.NewDecoder(resp.Body).Decode(&note)
	if note.ID != 123 {
		t.Errorf("unexpected id: %d", note.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	mockStore := &mockNoteStore{
		getByIDFn: func(ctx context.Context, id int) (model.Note, error) {
			return model.Note{}, store.ErrNotFound
		},
	}
	handler := NewNoteHandler(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/notes/999", nil)
	req = req.WithContext(setChiParam(req.Context(), "id", "999"))
	w := httptest.NewRecorder()
	handler.GetByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetByID_InvalidID(t *testing.T) {
	mockStore := &mockNoteStore{} // не должен вызываться
	handler := NewNoteHandler(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/notes/abc", nil)
	req = req.WithContext(setChiParam(req.Context(), "id", "abc"))
	w := httptest.NewRecorder()
	handler.GetByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// Вспомогательная функция для проброса параметров chi в контекст (без полноценного роутера).
func setChiParam(ctx context.Context, key, value string) context.Context {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}
