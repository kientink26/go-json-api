package application

import (
	"errors"
	"github.com/kientink26/go-json-api/cmd/api/helpers"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"github.com/kientink26/go-json-api/internal/data/postgresql"
	"github.com/kientink26/go-json-api/internal/validator"
	"net/http"
)

func (app *Application) getUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	p, err := app.Models.Permissions.GetAllForUser(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"permissions": p}, nil)
}

func (app *Application) addUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var p dto.Permissions
	err = helpers.ReadJSON(w, r, &p)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	dto.ValidatePermissions(v, p)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.Models.Permissions.AddForUser(id, p...)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, postgresql.ErrDuplicatePermission):
			v.AddError("permission", "a permission already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "permissions successfully added"}, nil)
}

func (app *Application) deleteUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var p dto.Permissions
	err = helpers.ReadJSON(w, r, &p)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	dto.ValidatePermissions(v, p)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.Models.Permissions.DeleteForUser(id, p...)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "permissions successfully deleted"}, nil)
}
