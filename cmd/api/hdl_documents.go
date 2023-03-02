package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"viadro_api/internal/data"
	"viadro_api/internal/logger"
	"viadro_api/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

// List all visible (public) documents
//
//	@Summary      List all visible (public) documents
//	@Description  List all visible (public) documents
//	@Tags         documents
//	@Produce      json
//	@Success      200  {object}   data.Document
//	@Failure      500  {string}  "Internal server error"
//	@Router       /documents [get]
func (app *application) listAllDocumentsHandler(w http.ResponseWriter, r *http.Request) {
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

	documents, metadata, err := app.data_access.Documents.GetAll(input.Title, input.Tags, input.Filters)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	responseSlice := []interface{}{}

	for _, document := range documents {
		doc := struct {
			ID          int64     `json:"document_id"`
			Title       string    `json:"title"`
			Link        string    `json:"link"`
			Tags        []string  `json:"tags"`
			Uploaded_at time.Time `json:"created_at"`
		}{
			ID:          document.Document_id,
			Title:       document.Title,
			Link:        document.Url_s3,
			Tags:        document.Tags,
			Uploaded_at: document.Uploaded_at,
		}

		responseSlice = append(responseSlice, doc)
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"metadata": metadata, "documents": responseSlice}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}
}

// List all user's documents
//
//	@Summary      List all user's documents
//	@Description  List all user's documents
//	@Tags         documents
//	@Produce      json
//	@Success      200  {object}   data.Document
//	@Failure      401  {string}  "Unauthorized"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /documents/my [get]
func (app *application) listUserDocumentsHandler(w http.ResponseWriter, r *http.Request) {
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

	user := app.contextGetUser(r)

	documents, metadata, err := app.data_access.Documents.GetUserAll(input.Title, input.Tags, input.Filters, user.ID)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	responseSlice := []interface{}{}

	for _, document := range documents {
		doc := struct {
			ID          int64     `json:"document_id"`
			Title       string    `json:"title"`
			Link        string    `json:"link"`
			Tags        []string  `json:"tags"`
			Uploaded_at time.Time `json:"created_at"`
			Is_hidden   bool      `json:"is_hidden"`
		}{
			ID:          document.Document_id,
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
		utils.ServerErrorResponse(w, r, err)
	}
}

// Add single document
//
//	@Summary      Add single document
//	@Description  Add single document
//	@Tags         document
//	@Accept       mpfd
//	@Produce      json
//	@Success      200  {object}   data.Document
//	@Failure      400  {string}  "Bad json reqest"
//	@Failure      401  {string}  "Unauthorized"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /document [post]
func (app *application) addDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Tags      []string `json:"tags"`
		Is_hidden bool     `json:"is_hidden"`
	}

	file, file_data, err := utils.ReadMultipartJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	defer file.Close()

	user := app.contextGetUser(r)

	document := &data.Document{
		User_id:   user.ID,
		Filetype:  ".pdf",
		Title:     file_data.Filename,
		Tags:      input.Tags,
		Is_hidden: input.Is_hidden,
	}

	uploader := manager.NewUploader(app.s3_client)
	res, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("viadro-api"),
		Key:    aws.String(document.Title),
		Body:   file,
		ACL:    "public-read",
	})
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	document.Url_s3 = res.Location

	err = app.data_access.Documents.Insert(document)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", document.Document_id))

	err = utils.WriteJSON(w, http.StatusCreated, utils.Wrap{"document": document}, headers)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

// Delete document
//
//	@Summary      Delete document
//	@Description  Delete document
//	@Tags         document
//	@Produce      json
//	@Success      200  {string}  "Successfully deleted"
//	@Failure      401  {string}  "Unauthorized"
//	@Failure      404  {string}  "Not found"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /document/:id [delete]
func (app *application) deleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		utils.NotFoundResponse(w, r)
		return
	}

	document, err := app.data_access.Documents.Get(id)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	if document.User_id != user.ID {
		utils.InvalidCredentialsResponse(w, r)
		return
	}

	output, err := app.s3_client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String("viadro-api"),
		Key:    aws.String(document.Title),
	})
	fmt.Println("OUTPUT: ", output)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = app.data_access.Documents.Delete(id)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"message": "document successfully deleted"}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

// Get document details
//
//	@Summary      Delete document
//	@Description  Delete document
//	@Tags         document
//	@Produce      json
//	@Success      200  {string}  data.Document
//	@Failure      401  {string}  "Unauthorized"
//	@Failure      404  {string}  "Not found"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /document/:id [get]
func (app *application) getDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		utils.NotFoundResponse(w, r)
		return
	}

	document, err := app.data_access.Documents.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			logger.LogError("Document does not exist", err) //? http.StatusNotFound - 404
			utils.NotFoundResponse(w, r)
		default:
			logger.LogError("failed getting document", err) //? http.StatusInternalServerError - 500
			utils.ServerErrorResponse(w, r, err)
		}
		return
	}

	user := app.contextGetUser(r)

	if document.Is_hidden && document.User_id != user.ID {
		utils.InvalidCredentialsResponse(w, r) //? http.StatusUnauthorized - 401
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"document": document}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

// Toggle document visibility
//
//	@Summary      Toggle document visibility
//	@Description  Toggle document visibility
//	@Tags         document
//	@Produce      json
//	@Success      200  {string}  "Successfully toggled visibility"
//	@Failure      401  {string}  "Unauthorized"
//	@Failure      404  {string}  "Not found"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /document/:id [patch]
func (app *application) toggleDocumentVisibilityHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		utils.NotFoundResponse(w, r)
		return
	}

	document, err := app.data_access.Documents.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			logger.LogError("Document does not exist", err) //? http.StatusNotFound - 404
			utils.NotFoundResponse(w, r)
		default:
			logger.LogError("failed getting document", err) //? http.StatusInternalServerError - 500
			utils.ServerErrorResponse(w, r, err)
		}
		return
	}

	user := app.contextGetUser(r)

	if document.User_id != user.ID {
		utils.InvalidCredentialsResponse(w, r)
		return
	}

	document, err = app.data_access.Documents.ToggleVisibility(id)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	doc := struct {
		ID          int64     `json:"document_id"`
		Title       string    `json:"title"`
		Link        string    `json:"link"`
		Tags        []string  `json:"tags"`
		Uploaded_at time.Time `json:"created_at"`
		Is_hidden   bool      `json:"is_hidden"`
	}{
		ID:          document.Document_id,
		Title:       document.Title,
		Link:        document.Url_s3,
		Tags:        document.Tags,
		Uploaded_at: document.Uploaded_at,
		Is_hidden:   document.Is_hidden,
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"document": doc}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}
