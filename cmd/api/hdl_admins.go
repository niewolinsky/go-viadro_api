package main

import (
	"errors"
	"net/http"
	"viadro_api/internal/data"
	"viadro_api/internal/logger"
	"viadro_api/utils"
)

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
func (app *application) toggleAdminGrant(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		utils.NotFoundResponse(w, r)
		return
	}

	user, err := app.data_access.Users.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			logger.LogError("user not found", err) //? http.StatusUnprocessableEntity - 422
			utils.FailedValidationResponse(w, r, map[string]string{"user not found": "true"})
		default:
			logger.LogError("internal error during retriving user by id", err) //? http.StatusInternalServerError - 500
			utils.ServerErrorResponse(w, r, err)
		}
		return
	}

	user.IsAdmin = !user.IsAdmin

	err = app.data_access.Users.Update(user)
	if err != nil {
		logger.LogError("failed updating user activated field", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"user": user}, nil)
	if err != nil {
		logger.LogError("failed to write response", err) //? http.StatusInternalServerError - 500
		utils.ServerErrorResponse(w, r, err)
	}
}

func (app *application) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.data_access.Users.GetAll()
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"users": users}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}
}
