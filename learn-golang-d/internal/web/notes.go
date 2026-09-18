package web

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"learn-golang-d/internal/store"
)

type NotesHandler struct {
	store *store.NoteStore
}

func NewNotesHandler(s *store.NoteStore) *NotesHandler {
	return &NotesHandler{store: s}
}

// List: GET /notes, publik (baca tidak butuh token, sama seperti B.18 di
// learn-golang-c - boundary publik/protected didefinisikan per rute, bukan
// global).
func (h *NotesHandler) List(c echo.Context) error {
	return c.JSON(http.StatusOK, h.store.List())
}

type createNotePayload struct {
	Title string `json:"title" form:"title" query:"title" validate:"required"`
	Body  string `json:"body" form:"body" query:"body"`
}

// Create (C.4 Bind menangani JSON/form/query lewat satu pemanggilan +
// C.5 Validate): POST /notes, diproteksi JWTMiddleware.
func (h *NotesHandler) Create(c echo.Context) error {
	var payload createNotePayload
	if err := c.Bind(&payload); err != nil {
		return err
	}
	if err := c.Validate(&payload); err != nil {
		return err
	}
	n := h.store.Create(payload.Title, payload.Body)
	return c.JSON(http.StatusCreated, n)
}

func (h *NotesHandler) Detail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id tidak valid")
	}
	n, ok := h.store.Get(id)
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "note tidak ditemukan")
	}
	return c.JSON(http.StatusOK, n)
}

func (h *NotesHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id tidak valid")
	}
	if !h.store.Delete(id) {
		return echo.NewHTTPError(http.StatusNotFound, "note tidak ditemukan")
	}
	return c.NoContent(http.StatusNoContent)
}
