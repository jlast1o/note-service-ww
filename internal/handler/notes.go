package handler

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"service/internal/model"
	"service/internal/store"
	"strconv"

	//_ "service/docs"

	"github.com/go-chi/chi"
)

type PaginatedResponse struct {
	Data       []model.Note `json:"data"`
	Page       int          `json:"page"`
	Size       int          `json:"size"`
	Total      int          `json:"total"`
	TotalPages int          `json:"total_pages"`
}

type NoteHandler struct {
	store store.NoteStorer
}

func NewNoteHandler(s store.NoteStorer) *NoteHandler {
	return &NoteHandler{store: s}
}

// Create создаёт новую заметку.
// @Summary      Создать заметку
// @Description  Создаёт новую заметку с заданным title и опциональным content.
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        note body      model.Note true "Данные заметки (обязательное поле title)"
// @Success      201  {object}  model.Note "Успешное создание"
// @Failure      400  {object}  map[string]string "Ошибка валидации"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка"
// @Router       /notes [post]

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if input.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}

	note, err := h.store.Create(r.Context(), input.Title, input.Content)

	if err != nil {
		log.Printf("ОШИБКА при создании заметки: %v", err)
		http.Error(w, `{"error":"could not create note"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

// GetAll возвращает все заметки.
// @Summary      Получить все заметки
// @Description  Возвращает список всех заметок, отсортированных по ID.
// @Tags         notes
// @Produce      json
// @Success      200  {array}   model.Note "Список заметок"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка"
// @Router       /notes [get]

func (h *NoteHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	// default values
	page := 1
	size := 20

	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			http.Error(w, `{"error":"invalid page"}`, http.StatusBadRequest)
			return
		}
	}

	if pageStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 1 {
			http.Error(w, `{"error":"invalid size"}`, http.StatusBadRequest)
			return
		}

		if size > 100 {
			size = 100
		}
	}

	offset := (page - 1) * size
	log.Printf("DEBUG limit=%d offset=%d", size, offset)
	notes, total, err := h.store.GetAll(r.Context(), size, offset)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch notes", "error", err)
		http.Error(w, `{"error":"could not fetch notes"}`, http.StatusInternalServerError)
		return
	}

	// Вычисляем totalPages (с защитой от деления на 0)
	totalPages := 0
	if size > 0 {
		totalPages = (total + size - 1) / size
	}

	// Обеспечиваем, что data не будет null
	if notes == nil {
		notes = make([]model.Note, 0)
	}

	resp := PaginatedResponse{
		Data:       notes,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}

// GetByID возвращает заметку по ID.
// @Summary      Получить заметку по ID
// @Description  Возвращает одну заметку по её идентификатору.
// @Tags         notes
// @Produce      json
// @Param        id   path      int true "ID заметки"
// @Success      200  {object}  model.Note "Найденная заметка"
// @Failure      400  {object}  map[string]string "Некорректный ID"
// @Failure      404  {object}  map[string]string "Заметка не найдена"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка"
// @Router       /notes/{id} [get]

func (h *NoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	note, err := h.store.GetByID(r.Context(), id)

	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, `{"error":"note not found"}`, http.StatusNotFound)
			return
		}

		http.Error(w, `{"error":"could not get note"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(note)

}
