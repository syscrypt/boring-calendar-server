package handler

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/syscrypt/boring-calendar/internal/service"
)

type Handler struct {
	router          *mux.Router
	db              *sql.DB
	calendarService *service.CalendarService
}

func NewHandler(db *sql.DB, calendarService *service.CalendarService) *Handler {
	router := mux.NewRouter()
	router.Path("/push").Methods(http.MethodPut, http.MethodOptions).HandlerFunc(calendarService.Create)
	router.Use(CorsMiddleware)

	return &Handler{
		router: router,
		db:     db,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
