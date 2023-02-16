package main

import (
	"net/http"
	"time"
	"viadro_api/internal/data"
	"viadro_api/internal/logger"
	"viadro_api/internal/mail"
	"viadro_api/utils"
)

func (app *application) userRegister(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Username:  input.Username,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = app.data_access.Users.Insert(user)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	token, err := app.data_access.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	data := map[string]interface{}{
		"activationToken": token.Plaintext,
		"userID":          user.ID,
	}

	email, err := mail.PrepareEmail(user.Email, "user_welcome.tmpl", data)
	if err != nil {
		logger.LogError("Failed to prepare email", err)
		utils.ServerErrorResponse(w, r, err)
		return
	}

	go func() {
		err = app.mail_client.DialAndSend(email)
		if err != nil {
			logger.LogError("Failed to send email", err)
			utils.ServerErrorResponse(w, r, err)
			return
		}
	}()

	err = utils.WriteJSON(w, http.StatusAccepted, utils.Wrap{"user": user}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

func (app *application) userActivate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}

	user, err := app.data_access.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	user.Activated = true

	app.data_access.Users.Update(user)

	app.data_access.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"user": user}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

func (app *application) userAuthenticate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}

	user, err := app.data_access.Users.GetByEmail(input.Email)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	if !match {
		utils.InvalidCredentialsResponse(w, r)
		return
	}

	token, err := app.data_access.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Wrap{"authentication_token": token}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
}

func (app *application) userLogout(w http.ResponseWriter, r *http.Request) {
}
