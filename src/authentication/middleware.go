package authentication

import (
	"context"
	"log"
	"net/http"
	"time"
)

type contextKey string
const UserKey contextKey = "user"

// Check if cookies contain valid jwt. Otherwise redirect to login.
// If valid jwt exists then adds user to request context with key "user".
func RequiredAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()	

		log.Printf("%s %s Unauthorized Request (no token)", r.Method, r.RequestURI)
		cookie, err := r.Cookie("token")
		if err != nil {
			log.Printf("%s %s Unauthorized Request (no token)", r.Method, r.RequestURI)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		log.Printf("Started parsing token (elapsed: %.3f ms)", float64(time.Since(start).Microseconds())/1000)
		claims, err := ParseToken(cookie.Value)
		log.Printf("Done parsing token (elapsed: %.3f ms)", float64(time.Since(start).Microseconds())/1000)
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
		ctx := context.WithValue(r.Context(), UserKey, user)
		log.Printf("Request finished (elapsed: %.3f ms)", float64(time.Since(start).Microseconds())/1000)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Check if cookies contain valid jwt. Adds user to request context with key "user"
// if valid jwt exists. Otherwise just executes next handler.
func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("Started parsing token (elapsed: %.3f ms)", float64(time.Since(start).Microseconds())/1000)
		cookie, err := r.Cookie("token")
		log.Printf("Done parsing token (elapsed: %.3f ms)", float64(time.Since(start).Microseconds())/1000)
		if err == nil {
			if claims, err := ParseToken(cookie.Value); err == nil {
				user := &User{
					UserID:   claims.UserID,
					Username: claims.Username,
					Email:    claims.Email,
				}
				ctx := context.WithValue(r.Context(), UserKey, user)
				r = r.WithContext(ctx)
			}
		}
		log.Printf("Request finished (elapsed: %.3f ms)", float64(time.Since(start).Microseconds())/1000)
		next.ServeHTTP(w, r)
	})
}
