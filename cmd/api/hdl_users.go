package main

import (
	"net/http"
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

	// app.background(func() {

	// 	data := map[string]interface{}{
	// 		"activationToken": token.Plaintext,
	// 		"userID":          user.ID,
	// 	}

	// 	err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
	// 	if err != nil {
	// 		app.logger.PrintError(err, nil)
	// 	}
	// })

	data := map[string]interface{}{
		"activationToken": "asdasdsa",
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

func (app *application) userLogout(w http.ResponseWriter, r *http.Request) {

}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {

}
