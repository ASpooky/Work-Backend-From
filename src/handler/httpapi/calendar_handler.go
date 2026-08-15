package httpapi

import (
	"net/http"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

type CalendarHandler struct {
	getCalendar *usecase.GetCalendarUsecase
}

func NewCalendarHandler(getCalendar *usecase.GetCalendarUsecase) *CalendarHandler {
	return &CalendarHandler{getCalendar: getCalendar}
}

func (h *CalendarHandler) Get(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")

	from, err := time.Parse(dateLayout, r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, "invalid from, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	to, err := time.Parse(dateLayout, r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, "invalid to, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	got, err := h.getCalendar.Execute(usecase.GetCalendarInput{
		WorkspaceID: workspaceID,
		From:        from,
		To:          to,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, got)
}
