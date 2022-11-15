package helpers

import (
	"context"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"net/http"
)

// Define a custom contextKey type, with the underlying type string.
type contextKey string

// We'll use this constant as the key for getting and setting user information
// in the request context.
const userContextKey = contextKey("user")

// The ContextSetUser() method returns a new copy of the request with the provided
// User struct added to the context.
func ContextSetUser(r *http.Request, user *dto.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// The contextSetUser() retrieves the User struct from the request context
func ContextGetUser(r *http.Request) *dto.User {
	user, ok := r.Context().Value(userContextKey).(*dto.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
