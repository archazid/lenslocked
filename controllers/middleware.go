package controllers

import (
	"net/http"

	"archazid.io/lenslocked/context"
	"archazid.io/lenslocked/models"
)

type UserMiddleware struct {
	SessionService *models.SessionService
}

func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the cookie. If it returns error, proceed with the request.
		// The goal isn't to limit access, but to sets the user in the context.
		token, err := readCookie(r, CookieSession)
		if err != nil {
			// Cannot lookup the user with no cookie,
			// so proceed without a user being set.
			next.ServeHTTP(w, r)
			return
		}

		// If we have a token, try to lookup the user with that token.
		user, err := umw.SessionService.User(token)
		if err != nil {
			// Invalid or expired token. Proceed without a user being set.
			next.ServeHTTP(w, r)
			return
		}

		// Store the user in the context.
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		// Get a request that uses our new context.
		r = r.WithContext(ctx)
		// Call the handler with the updated request.
		next.ServeHTTP(w, r)
	})
}

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
