package application

import (
	"errors"
	"fmt"
	"github.com/kientink26/go-json-api/cmd/api/helpers"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"github.com/kientink26/go-json-api/internal/data/postgresql"
	"github.com/kientink26/go-json-api/internal/validator"
	"net/http"
)

func (app *Application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		dto.Filters
	}
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	input.Title = helpers.ReadString(qs, "title", "")
	input.Genres = helpers.ReadCSV(qs, "genres", []string{})
	// We pass the validator instance as the final argument here.
	input.Page = helpers.ReadInt(qs, "page", 1, v)
	input.PageSize = helpers.ReadInt(qs, "page_size", 20, v)
	// Extract the sort query string value, falling back to "id" if it is not provided
	// by the client (which will imply a ascending sort on movie ID).
	input.Sort = helpers.ReadString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	if dto.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	movies, metadata, err := app.Models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
func (app *Application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	movie, err := app.Models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *Application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string      `json:"title"`
		Year    int32       `json:"year"`
		Runtime dto.Runtime `json:"runtime"`
		Genres  []string    `json:"genres"`
	}
	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	movie := &dto.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}
	v := validator.New()
	dto.ValidateMovie(v, movie)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.Models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))
	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *Application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	movie, err := app.Models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	var input struct {
		Title   *string      `json:"title"` // This will be nil if there is no corresponding key in the JSON
		Year    *int32       `json:"year"`
		Runtime *dto.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	err = helpers.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres // Note that we don't need to dereference a slice.
	}
	v := validator.New()
	dto.ValidateMovie(v, movie)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.Models.Movies.Update(movie)
	// Intercept any ErrEditConflict error and call the new editConflictResponse()
	// helper.
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *Application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the movie from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.Models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return a 200 OK status code along with a success message.
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
