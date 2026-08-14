package auth

import (
	"net/http"

	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
)

// TODO Add a cache for the "me" checks to reduce database calls

// NewAccountMiddleware returns a middleware that checks if the user has a full account and returns 403 if he doesn't
func NewAccountMiddleware(store db.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := GetIdentityFromContext(r.Context())
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			user, err := store.GetUserById(r.Context(), identity.Id)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			identity.HasSidequestAccount = true
			identity.Username = user.Username
			ctx := SetIdentityToContext(r.Context(), identity)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
