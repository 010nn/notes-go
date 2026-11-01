package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"notes-go/internal/config"
	"notes-go/internal/repository/postgres"
	"notes-go/internal/service"
	"notes-go/internal/transport"
	"notes-go/internal/transport/handler"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	// load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		return
	}

	// connect DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		slog.Error("connect db", "err", err)
		return
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		slog.Error("unavalible db", "err", err)
		return
	}

	regexUUID := "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"

	staticHandler := handler.NewStaticHandler()
	JWTManager := transport.NewJWTManager(cfg.JwtSecret, time.Duration(cfg.JwtAccessExpirationMinute)*time.Minute, 0)

	// Note
	noteRepository := postgres.NewNoteRepository(dbPool)
	noteService := service.NewNoteService(noteRepository)
	notesHandler := handler.NewNotesHandler(noteService)

	// Auth
	userRepository := postgres.NewUserRepository(dbPool)
	userService := service.NewUserService(userRepository, JWTManager)
	authHandler := handler.NewAuthHandler(userService, JWTManager)

	middleware := transport.NewMiddleware(JWTManager)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(time.Minute))
	r.Use(chimiddleware.Logger)

	// Static files
	// fileServer := http.FileServer(handler.NewNeuteredFileSystem("./static"))
	// r.Group(func(r chi.Router) {
	// 	r.Use(middleware.StaticSmartCache)
	// 	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))
	// })

	// API routing
	r.Route("/api", func(r chi.Router) {
		// Public
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Post("/register", authHandler.Register)
		})

		// Private
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthRequired)
			r.Get("/auth/me", authHandler.Me)
		})

		r.Group(func(r chi.Router) {
			apiNoteId := fmt.Sprintf("/notes/{id:%s}", regexUUID)
			r.Use(middleware.AuthRequired)
			r.Get("/notes", notesHandler.List)
			r.Post(apiNoteId, notesHandler.Update)
			r.Delete(apiNoteId, notesHandler.Delete)
			r.Post("/notes/reorder", notesHandler.Reorder)
		})
	})

	// Pages routing
	r.Get("/login", staticHandler.Auth)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRequired)
		r.Get("/", staticHandler.Notes)
	})

	// Server
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	slog.Info("Server running on http://127.0.0.1:" + cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("run server", "err", err)
		return
	}

}
