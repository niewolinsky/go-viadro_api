package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
	"viadro_api/internal/data"
	"viadro_api/internal/logger"
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
//	@Router       /user/register [post]
func (app *application) userRegister(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		logger.LogError("malformed json request", err) //? http.StatusBadRequest - 400
		utils.BadRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Username:  input.Username,
		Email:     input.Email,
		Activated: false,
		IsAdmin:   false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		logger.LogError("failed to generate password hash", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = app.data_access.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			logger.LogError("account with this email already exists", err) //? http.StatusUnprocessableEntity - 422
			utils.FailedValidationResponse(w, r, map[string]string{"duplicate email": "true"})
		default:
			logger.LogError("failed to create new user", err) //? http.StatusInternalServerError - 500
			utils.ServerErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.data_access.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		logger.LogError("failed creating activation token", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
		return
	}

	fmt.Println("Activation token instead of email: ", token.Plaintext)
	//!uncomment later
	// data := map[string]interface{}{
	// 	"activationToken": token.Plaintext,
	// 	"userID":          user.ID,
	// }

	// email, err := mail.PrepareEmail(user.Email, "user_welcome.html", data)
	// if err != nil {
	// 	logger.LogError("failed to prepare email", err) //? http.StatusInternalServerError - 500
	// 	utils.ServerErrorResponse(w, r, err)
	// 	return
	// }

	// go func() {
	// 	err = app.mail_client.DialAndSend(email)
	// 	if err != nil {
	// 		logger.LogError("failed to send email", err) //? http.StatusInternalServerError - 500
	// 		utils.ServerErrorResponse(w, r, err)
	// 		return
	// 	}
	// }()

	err = utils.WriteJSON(w, http.StatusAccepted, utils.Wrap{"user": user}, nil)
	if err != nil {
		logger.LogError("failed to write response", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
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
func (app *application) userActivate(w http.ResponseWriter, r *http.Request) {
	input := struct {
		TokenPlaintext string `json:"token"`
	}{}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		logger.LogError("malformed request json", err) //? http.StatusBadRequest - 400
		utils.BadRequestResponse(w, r, err)
		return
	}

	user, err := app.data_access.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			logger.LogError("invalid or expired token", err) //? http.StatusUnprocessableEntity - 422
			utils.FailedValidationResponse(w, r, map[string]string{"invalid or expired token": "true"})
		default:
			logger.LogError("failed getting token for user", err) //? http.StatusInternalServerError - 500
			utils.ServerErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.data_access.Users.Update(user)
	if err != nil {
		logger.LogError("failed updating user activated field", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = app.data_access.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		logger.LogError("failed deleting activation token for user", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"user": user}, nil)
	if err != nil {
		logger.LogError("failed to write response", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
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
func (app *application) userAuthenticate(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		logger.LogError("malformed request json", err) //? http.StatusBadRequest - 400
		utils.BadRequestResponse(w, r, err)
		return
	}

	user, err := app.data_access.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			logger.LogError("invalid credidentals, user not found", err) //? http.StatusUnauthorized - 401
			utils.InvalidCredentialsResponse(w, r)
		default:
			logger.LogError("failed getting token for user", err) //? http.StatusInternalServerError - 500
			utils.ServerErrorResponse(w, r, err)
		}
		return
	}

	_, err = user.Password.Matches(input.Password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrBadPassword):
			logger.LogError("invalid credidentals, wrong password", err) //? http.StatusUnauthorized - 401
			utils.InvalidCredentialsResponse(w, r)
		default:
			logger.LogError("failed comparing passwords", err) //? http.StatusInternalServerError - 500
			utils.ServerErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.data_access.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		logger.LogError("failed creating authentication token", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Wrap{"authentication_token": token}, nil)
	if err != nil {
		logger.LogError("failed to write response", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
	}
}
