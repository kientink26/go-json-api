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

func (app *Application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string
		Email string
		dto.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	input.Name = helpers.ReadString(qs, "name", "")
	input.Email = helpers.ReadString(qs, "email", "")
	// We pass the validator instance as the final argument here.
	input.Page = helpers.ReadInt(qs, "page", 1, v)
	input.PageSize = helpers.ReadInt(qs, "page_size", 20, v)
	// Extract the sort query string value, falling back to "id" if it is not provided
	// by the client.
	input.Sort = helpers.ReadString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "name", "email", "created_at", "-id", "-name", "-email", "-created_at"}
	if dto.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	users, metadata, err := app.Models.Users.GetAll(input.Name, input.Email, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"users": users, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an anonymous struct to hold the expected data from the request body.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// Parse the request body into the anonymous struct.
	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Copy the data from the request body into a new User struct. We
	// set the Activated field to false.
	user := &dto.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	// Use the Password.Set() method to generate and store the hashed and plaintext
	// passwords.
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()
	if dto.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.Models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Add the "comments:write" permission for the new user.
	err = app.Models.Permissions.AddForUser(user.ID, dto.CommentsWrite)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	token, err := app.Models.Tokens.New(user.ID, 3*24*time.Hour, dto.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}
		// Send the welcome email, passing in the map above as dynamic data.
		err = app.Mailer.Send(user.Email, "user_welcome.gohtml", data)
		if err != nil {
			app.logError(err)
		}

	})
	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Parse the plaintext activation token from the request body.
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err = helpers.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Validate the plaintext token provided by the client.
	v := validator.New()
	if dto.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Retrieve the details of the user associated with the token using the
	// GetForToken() method
	user, err := app.Models.Users.GetForToken(dto.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Check the ID of the user associated with the token
	if user.ID != id {
		v.AddError("token", "incorrect activation token for this user")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Update the user's activation status.
	user.Activated = true
	err = app.Models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// If everything went successfully, then we delete all activation tokens for the
	// user.
	err = app.Models.Tokens.DeleteAllForUser(dto.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send the updated user details to the client in a JSON response.
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
