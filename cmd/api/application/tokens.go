package application

import (
	"errors"
	"github.com/kientink26/go-json-api/cmd/api/helpers"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"github.com/kientink26/go-json-api/internal/data/postgresql"
	"github.com/kientink26/go-json-api/internal/validator"
	"net/http"
	"time"
)

func (app *Application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the email and password from the request body.
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Validate the email and password provided by the client.
	v := validator.New()
	dto.ValidateEmail(v, input.Email)
	dto.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.Models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Check if the provided password matches the actual password for the user.
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	// Otherwise, if the password is correct, we generate a new token with a 24-hour
	// expiry time and the scope 'authentication'.
	token, err := app.Models.Tokens.New(user.ID, 24*time.Hour, dto.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Encode the token to JSON and send it in the response along with a 201 Created
	// status code.
	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
