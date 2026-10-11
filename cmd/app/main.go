package main

import (
	"fmt"
	"log"
	"net/http"
	"notes-go/internal/config"
	"notes-go/internal/repository/json"
	"notes-go/internal/repository/memory"
	"notes-go/internal/service"
	"notes-go/internal/transport"
	"notes-go/internal/transport/handler"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {

	fileServer := http.FileServer(handler.NewNeuteredFileSystem("./static"))
	regexUUID := "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"

	staticHandler := handler.NewStaticHandler()

	// Note
	noteRepository := json.NewNoteJSONRepository()
	noteService := service.NewNoteService(noteRepository)
	notesHandler := handler.NewNotesHandler(noteService)

	// Auth
	userRepository := memory.NewUserMemoryRepository()
	userService := service.NewUserService(userRepository)
	authHandler := handler.NewAuthHandler(userService)

	middleware := transport.NewMiddleware()

	r := chi.NewRouter()
	r.Use(middleware.Logging)

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
			r.Get(apiNoteId, notesHandler.Get)
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

	log.Printf("Server running on http://127.0.0.1:%s\n", config.PORT)
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}

}
