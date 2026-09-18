package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"learn-golang-d/internal/store"
)

type HabitsHandler struct {
	store *store.HabitStore
}

func NewHabitsHandler(s *store.HabitStore) *HabitsHandler {
	return &HabitsHandler{store: s}
}

type habitView struct {
	store.Habit
	CheckedInToday bool `json:"checked_in_today"`
	Streak         int  `json:"streak"`
}

func (h *HabitsHandler) view(hb store.Habit) habitView {
	today := time.Now().Format("2006-01-02")
	return habitView{
		Habit:          hb,
		CheckedInToday: h.store.IsCheckedIn(hb.ID, today),
		Streak:         h.store.Streak(hb.ID),
	}
}

func (h *HabitsHandler) List(c echo.Context) error {
	habits := h.store.List()
	out := make([]habitView, 0, len(habits))
	for _, hb := range habits {
		out = append(out, h.view(hb))
	}
	return c.JSON(http.StatusOK, out)
}

type createHabitPayload struct {
	Name string `json:"name" form:"name" query:"name" validate:"required"`
}

func (h *HabitsHandler) Create(c echo.Context) error {
	var payload createHabitPayload
	if err := c.Bind(&payload); err != nil {
		return err
	}
	if err := c.Validate(&payload); err != nil {
		return err
	}
	hb := h.store.Create(payload.Name)
	return c.JSON(http.StatusCreated, h.view(hb))
}

func (h *HabitsHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id tidak valid")
	}
	if !h.store.Delete(id) {
		return echo.NewHTTPError(http.StatusNotFound, "habit tidak ditemukan")
	}
	return c.NoContent(http.StatusNoContent)
}

// Checkin: POST /habits/{id}/checkin, toggle absensi hari ini.
func (h *HabitsHandler) Checkin(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id tidak valid")
	}
	today := time.Now().Format("2006-01-02")
	done, ok := h.store.ToggleCheckIn(id, today)
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "habit tidak ditemukan")
	}
	hb, _ := h.store.Get(id)
	return c.JSON(http.StatusOK, map[string]any{
		"id":     hb.ID,
		"date":   today,
		"done":   done,
		"streak": h.store.Streak(id),
	})
}
