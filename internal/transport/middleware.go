package transport

import (
	"net/http"
	"notes-go/pkg/jwt"
	"strings"
)

type Middleware struct {
	jwtManager *jwt.JWTManager
}

func NewMiddleware(jwtManager *jwt.JWTManager) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
	}
}

func (m *Middleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		token := cookie.Value

		claims, err := m.jwtManager.Verify(token)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := WithUser(r.Context(), claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) StaticSmartCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, ".css") {
			w.Header().Set("Cache-Control", "public, max-age=900, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		}

		next.ServeHTTP(w, r)
	})
}
