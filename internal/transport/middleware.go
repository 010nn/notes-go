package transport

import (
	"fmt"
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

		cookie, err := r.Cookie("SUPERTOKEN")
		if err != nil {
			// http.Error(w, "invalid token", http.StatusBadRequest)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		log.Printf("token: %s", cookie.Value)

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Println()
		log.Printf("%s %s %s %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
		log.Printf("Header: %v", r.Header)
		fmt.Println()

		next.ServeHTTP(w, r)
	})
}
