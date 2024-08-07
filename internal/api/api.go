package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GuiCezaF/Go-React/internal/store/pgstore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx"
)

type apiHandler struct {
	q        *pgstore.Queries
	r        *chi.Mux
	upgrader websocket.Upgrader
}

func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func NewHandler(q *pgstore.Queries) http.Handler {
	a := apiHandler{
		q:        q,
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/subscribe/{room_id}", a.handleSubscribe)

	r.Route("/api", func(r chi.Router) {
		r.Route("/rooms", func(r chi.Router) {
			r.Post("/", a.handleCreateRoom)
			r.Get("/", a.handleGetRooms)

			r.Route("/{room_id}/messages", func(r chi.Router) {
				r.Get("/", a.handleCreateRoomMessages)
				r.Post("/", a.handleGetRoomMessages)

				r.Route("/{message_id}", func(r chi.Router) {
					r.Get("/", a.handleGetRoomMessage)
					r.Patch("/react", a.hadleReactToMessage)
					r.Delete("/react", a.hadleRemoveReactFromMessage)
					r.Patch("/answer", a.hadleMarkMessageAsAnswered)
				})
			})
		})
	})

	a.r = r

	return a
}

func (h apiHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	rawuRoomId := chi.URLParam(r, "room_id")
	roomId, err := uuid.Parse(rawuRoomId)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	_, err = h.q.GetRoom(r.Context(), roomId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("failure to upgrade connection", "error", err)
		http.Error(w, "failed to upgrade to ws connection", http.StatusBadRequest)
		return
	}

	defer c.Close()

}
func (h apiHandler) handleCreateRoom(w http.ResponseWriter, r *http.Request)            {}
func (h apiHandler) handleGetRooms(w http.ResponseWriter, r *http.Request)              {}
func (h apiHandler) handleCreateRoomMessages(w http.ResponseWriter, r *http.Request)    {}
func (h apiHandler) handleGetRoomMessages(w http.ResponseWriter, r *http.Request)       {}
func (h apiHandler) handleGetRoomMessage(w http.ResponseWriter, r *http.Request)        {}
func (h apiHandler) hadleReactToMessage(w http.ResponseWriter, r *http.Request)         {}
func (h apiHandler) hadleRemoveReactFromMessage(w http.ResponseWriter, r *http.Request) {}
func (h apiHandler) hadleMarkMessageAsAnswered(w http.ResponseWriter, r *http.Request)  {}
