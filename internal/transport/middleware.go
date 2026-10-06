package transport

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Middleware struct {
}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("token")
		if err != nil {
			// http.Error(w, "invalid token", http.StatusBadRequest)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		token := cookie.Value

		claims, err := ParseJWSToken(token)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		log.Printf("token: %s\tname: %s", token, claims.Username)

		ctx := context.WithValue(r.Context(), "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("\nLog: %s %s %s %v\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))

		next.ServeHTTP(w, r)
	})
}
