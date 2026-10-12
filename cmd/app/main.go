package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"notes-go/internal/config"
	"notes-go/internal/repository/json"
	"notes-go/internal/repository/memory"
	"notes-go/internal/service"
	"notes-go/internal/transport"
	"notes-go/internal/transport/handler"
	"notes-go/pkg/jwt"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {

	fileServer := http.FileServer(handler.NewNeuteredFileSystem("./static"))
	regexUUID := "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"

	staticHandler := handler.NewStaticHandler()
	JWTManager := jwt.NewJWTManager(config.JWT_SECRET, time.Hour*24, 0)

	// Note
	noteRepository, err := json.NewNoteJSONRepository()
	if err != nil {
		slog.Error("note repo: ", "err", err)
		return
	}
	noteService := service.NewNoteService(noteRepository)
	notesHandler := handler.NewNotesHandler(noteService)

	// Auth
	userRepository := memory.NewUserMemoryRepository()
	userService := service.NewUserService(userRepository, JWTManager)
	authHandler := handler.NewAuthHandler(userService)

	middleware := transport.NewMiddleware(JWTManager)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(time.Minute))
	r.Use(chimiddleware.Logger)

	// Static files
	r.Group(func(r chi.Router) {
		r.Use(middleware.StaticSmartCache)
		r.Handle("/static/*", http.StripPrefix("/static/", fileServer))
	})

	// Public routing
	r.Get("/login", staticHandler.Auth)
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Post("/register", authHandler.Register)
		})

		r.Group(func(r chi.Router) {
			apiNoteId := fmt.Sprintf("/notes/{id:%s}", regexUUID)
			r.Use(middleware.AuthRequired)
			r.Get("/notes", notesHandler.List)
			r.Post(apiNoteId, notesHandler.Update)
			r.Delete(apiNoteId, notesHandler.Delete)
		})
	})

	// Authorizated routing
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRequired)
		r.Get("/", staticHandler.Notes)
	})

	// Server
	server := &http.Server{
		Addr:              ":" + config.PORT,
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	slog.Info("Server running on http://127.0.0.1:" + config.PORT)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("run server", "err", err)
		return
	}

}
