package application

import (
	"errors"
	"fmt"
	"github.com/kientink26/go-json-api/cmd/api/helpers"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"github.com/kientink26/go-json-api/internal/data/postgresql"
	"github.com/kientink26/go-json-api/internal/validator"
	"net/http"
	"strings"
)

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been
				// sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		// Retrieve the value of the Authorization header from the request. This will
		// return the empty string "" if there is no such header found.
		authorizationHeader := r.Header.Get("Authorization")
		// If there is no Authorization header found, add the AnonymousUser to the request context.
		// Then we call the next handler in the chain and return without executing any of the
		// code below.
		if authorizationHeader == "" {
			r = helpers.ContextSetUser(r, dto.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>"
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		token := headerParts[1]
		v := validator.New()
		if dto.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		user, err := app.Models.Users.GetForToken(dto.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, postgresql.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		// Call the contextSetUser() helper to add the user information to the request
		// context.
		r = helpers.ContextSetUser(r, user)
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// Create a new requireAuthenticatedUser() middleware to check that a user is not
// anonymous.
func (app *Application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.ContextGetUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Checks that a user is both authenticated and activated.
func (app *Application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// Rather than returning this http.HandlerFunc we assign it to the variable fn.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := helpers.ContextGetUser(r)
		// Check that a user is activated.
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	// Wrap fn with the requireAuthenticatedUser() middleware before returning it.
	return app.requireAuthenticatedUser(fn)
}

func (app *Application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user from the request context.
		user := helpers.ContextGetUser(r)
		// Get the slice of permissions for the user.
		permissions, err := app.Models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// Check if the slice includes the required permission. If it doesn't, then
		// return a 403 Forbidden response.
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}
		// Otherwise they have the required permission so we call the next handler in
		// the chain.
		next.ServeHTTP(w, r)
	}
	// Wrap this with the requireActivatedUser() middleware before returning it.
	return app.requireActivatedUser(fn)
}

func (app *Application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")
		if origin != "" && origin == app.Config.Cors.TrustedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// Check if the request has the HTTP method OPTIONS and contains the
			// "Access-Control-Request-Method" header. If it does, then we treat
			// it as a preflight request.
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				// Set the necessary preflight response headers
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				// Write the headers along with a 200 OK status and return from
				// the middleware with no further action.
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
