package main

import (
	"errors"
	"net/http"
	"time"
	"viadro_api/internal/data"
	"viadro_api/internal/mail"
	"viadro_api/utils"
)

// Register a new user
//
//	@Summary      Register a new user
//	@Description  Register a new user
//	@Tags         user
//	@Accept      json
//	@Produce      json
//	@Success      202  {object}   data.User
//	@Failure      400  {string}  "Bad json request"
//	@Failure      422  {string}  "User exists"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /user [post]
func (app *application) userRegisterHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err) //? http.StatusBadRequest - 400
		return
	}

	user := &data.User{
		Username:  input.Username,
		Email:     input.Email,
		Activated: false,
		Is_admin:  false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	err = app.data_access.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			utils.FailedValidationResponse(w, r, map[string]string{"duplicate email": "true"}) //? http.StatusUnprocessableEntity - 422
		default:
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		}
		return
	}

	token, err := app.data_access.Tokens.New(user.User_id, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	data := map[string]interface{}{
		"activation_token": token.Plaintext,
		"user_id":          user.User_id,
	}

	email, err := mail.PrepareEmail(user.Email, "user_welcome.html", data)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	go func() {
		err = app.mail_client.DialAndSend(email)
		if err != nil {
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
			return
		}
	}()

	err = utils.WriteJSON(w, http.StatusAccepted, utils.Wrap{"user": user}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}
}

// Activate user account
//
//	@Summary      Activate user account
//	@Description  Activate user account
//	@Tags         user
//	@Accept      json
//	@Produce      json
//	@Success      200  {string}  "User activated"
//	@Failure      400  {string}  "Bad json request"
//	@Failure      422  {string}  "Invalid or expired token"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /user/activate [put]
func (app *application) userActivateHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		TokenPlaintext string `json:"token"`
	}{}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err) //? http.StatusBadRequest - 400
		return
	}

	user, err := app.data_access.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			utils.FailedValidationResponse(w, r, map[string]string{"invalid or expired token": "true"}) //? http.StatusUnprocessableEntity - 422
		default:
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		}
		return
	}

	user.Activated = true

	err = app.data_access.Users.Update(user)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	err = app.data_access.Tokens.DeleteAllForUser(data.ScopeActivation, user.User_id)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"user": user}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}
}

// Authenticate (login) user
//
//	@Summary      Authenticate (login) user
//	@Description  Authenticate (login) user
//	@Tags         user
//	@Accept      json
//	@Produce      json
//	@Success      201  {string}  "User authenticated"
//	@Failure      400  {string}  "Bad json request"
//	@Failure      401  {string}  "Bad credentials"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /user/authenticate [put]
func (app *application) userAuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err) //? http.StatusBadRequest - 400
		return
	}

	user, err := app.data_access.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			utils.InvalidCredentialsResponse(w, r) //? http.StatusUnauthorized - 401
		default:
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		}
		return
	}

	_, err = user.Password.Matches(input.Password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrBadPassword):
			utils.InvalidCredentialsResponse(w, r) //? http.StatusUnauthorized - 401
		default:
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		}
		return
	}

	token, err := app.data_access.Tokens.New(user.User_id, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Wrap{"authentication_token": token}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}
}

// Delete (deactivate) user
//
//	@Summary      Delete (deactivate) user
//	@Description  Delete (deactivate) user
//	@Tags         user
//	@Produce      json
//	@Success      201  {string}  "User authenticated"
//	@Failure      400  {string}  "Bad json request"
//	@Failure      401  {string}  "Bad credentials"
//	@Failure      404  {string}  "User not found"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /user [delete]
func (app *application) userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		utils.NotFoundResponse(w, r) //? http.StatusNotFound - 404
		return
	}

	user, err := app.data_access.Users.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			utils.NotFoundResponse(w, r) //? http.StatusNotFound - 404
		default:
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		}
		return
	}

	user_ctx := app.contextGetUser(r)
	if user.User_id != user_ctx.User_id && !user_ctx.Is_admin {
		utils.InvalidCredentialsResponse(w, r) //? http.StatusUnauthorized - 401
		return
	}

	err = app.data_access.Users.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			utils.NotFoundResponse(w, r) //? http.StatusNotFound - 404
		default:
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, nil, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}
}
