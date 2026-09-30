package main

import (
	"log"
	"net/http"
	"notes-go/internal/config"
	"notes-go/internal/repository"
	"notes-go/internal/service"
	"notes-go/internal/transport"
	"notes-go/internal/transport/handler"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func main() {

	fileServer := http.FileServer(http.Dir("./static"))

	staticHandler := handler.NewStaticHandler()
	notesHandler := handler.NewNotesHandler()

	// Auth
	sessionKey := []byte("super-puper-secret-key-12345")
	store := sessions.NewCookieStore(sessionKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 3, // 3 часа
		HttpOnly: true,
		Secure:   false, // Включить, если используется HTTPS
	}
	userRepository := repository.NewUsersRepository()
	userService := service.NewUserService(userRepository)
	authHandler := handler.NewAuthHandler(userService, store)

	middleware := transport.NewMiddleware(store)

	r := mux.NewRouter()
	r.Use(middleware.Logging)

	// Public routing
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
	r.HandleFunc("/login", staticHandler.Auth).Methods("GET")

	publicApi := r.PathPrefix("/api").Subrouter()
	publicApi.HandleFunc("/login", authHandler.Login).Methods("POST")
	publicApi.HandleFunc("/register", authHandler.Register).Methods("POST")

	// Authorizated routing
	mainPage := r.PathPrefix("/").Subrouter()
	mainPage.Use(middleware.AuthRequired)
	mainPage.HandleFunc("/", staticHandler.Notes).Methods("GET")

	privateApi := r.PathPrefix("/api").Subrouter()
	privateApi.Use(middleware.AuthRequired)
	privateApi.HandleFunc("/note/{id:[0-9]+}", notesHandler.GetNote).Methods("GET")

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
