package authentication

import (
	"context"
	"log"
	"net/http"
)

// Check if cookies contain valid jwt. Otherwise redirect to login.
// If valid jwt exists then adds user to request context with key "user".
func RequiredAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			log.Printf("%s %s Unauthorized Request (no token)", r.Method, r.RequestURI)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		claims, err := ParseToken(cookie.Value)
		if err != nil {
			log.Printf("%s %s Unauthorized Request (invalid token): %s", r.Method, r.RequestURI, err.Error())
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		user := &User{
			UserID:   claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Check if cookies contain valid jwt. Adds user to request context with key "user"
// if valid jwt exists. Otherwise just executes next handler.
func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err == nil {
			if claims, err := ParseToken(cookie.Value); err == nil {
				user := &User{
					UserID:   claims.UserID,
					Username: claims.Username,
					Email:    claims.Email,
				}
				ctx := context.WithValue(r.Context(), "user", user)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
