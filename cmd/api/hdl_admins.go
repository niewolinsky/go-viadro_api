package main

import (
	"errors"
	"net/http"
	"time"
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
func (app *application) toggleAdminGrantHandler(w http.ResponseWriter, r *http.Request) {
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

	user.Is_admin = !user.Is_admin

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

func (app *application) getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
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

func (app *application) getAllDocumentsAdminHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	input := struct {
		Title string
		Tags  []string
		data.Filters
	}{}

	input.Title = utils.ReadStringParam(qs, "title", "")
	input.Tags = utils.ReadCSVParam(qs, "tags", []string{})
	input.Filters.Page = utils.ReadIntParam(qs, "page", 1)
	input.Filters.PageSize = utils.ReadIntParam(qs, "page_size", 20)
	input.Filters.Sort = utils.ReadStringParam(qs, "sort", "document_id")
	input.Filters.SortSafelist = []string{"document_id", "-document_id"}

	documents, metadata, err := app.data_access.Documents.GetAllAdmin(input.Title, input.Tags, input.Filters)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	responseSlice := []interface{}{}

	for _, document := range documents {
		doc := struct {
			ID          int       `json:"document_id"`
			User_id     int       `json:"user_id"`
			Title       string    `json:"title"`
			Link        string    `json:"link"`
			Tags        []string  `json:"tags"`
			Uploaded_at time.Time `json:"created_at"`
			Is_hidden   bool      `json:"is_hidden"`
		}{
			ID:          document.Document_id,
			User_id:     document.User_id,
			Title:       document.Title,
			Link:        document.Url_s3,
			Tags:        document.Tags,
			Uploaded_at: document.Uploaded_at,
			Is_hidden:   document.Is_hidden,
		}

		responseSlice = append(responseSlice, doc)
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"metadata": metadata, "documents": responseSlice}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}
}
