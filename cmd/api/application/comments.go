package application

import (
	"errors"
	"github.com/kientink26/go-json-api/cmd/api/helpers"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"github.com/kientink26/go-json-api/internal/data/postgresql"
	"github.com/kientink26/go-json-api/internal/validator"
	"net/http"
)

func (app *Application) listCommentsHandler(w http.ResponseWriter, r *http.Request) {
	movieID, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	filter := dto.Filters{}
	v := validator.New()
	qs := r.URL.Query()
	filter.Sort = helpers.ReadString(qs, "sort", "id")
	filter.SortSafelist = []string{"id", "created_at", "-id", "-created_at"}
	if dto.ValidateSortQuery(v, filter); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	comments, metadata, err := app.Models.Comments.GetAllForMovie(movieID, filter)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"comments": comments, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	movieID, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var input struct {
		Body string `json:"body"`
	}
	err = helpers.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	comment := &dto.Comment{
		Body: input.Body,
	}
	v := validator.New()
	if dto.ValidateComment(v, comment); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.Models.Comments.Insert(comment, helpers.ContextGetUser(r).ID, movieID)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
