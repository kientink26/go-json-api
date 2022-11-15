package application

import (
	"github.com/julienschmidt/httprouter"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"net/http"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission(dto.MoviesWrite, app.createMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission(dto.MoviesWrite, app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission(dto.MoviesWrite, app.deleteMovieHandler))

	router.HandlerFunc(http.MethodGet, "/v1/movies/:id/comments", app.requireActivatedUser(app.listCommentsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies/:id/comments", app.requirePermission(dto.CommentsWrite, app.createCommentHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/:id/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.HandlerFunc(http.MethodGet, "/v1/users", app.requirePermission(dto.UsersRead, app.listUsersHandler))
	router.HandlerFunc(http.MethodGet, "/v1/users/:id/permissions", app.requirePermission(dto.PermissionsRead, app.getUserPermissionsHandler))
	router.HandlerFunc(http.MethodPut, "/v1/users/:id/permissions", app.requirePermission(dto.PermissionsWrite, app.addUserPermissionsHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/users/:id/permissions", app.requirePermission(dto.PermissionsWrite, app.deleteUserPermissionsHandler))

	return app.recoverPanic(app.enableCORS(app.authenticate(router)))
}
