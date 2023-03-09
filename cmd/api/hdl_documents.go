package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
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
//	@Tags         document
//	@Produce      json
//	@Success      200  {object}   data.Document
//	@Failure      500  {string}  "Internal server error"
//	@Router       /documents [get]
func (app *application) getAllDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	if len(qs) == 0 {
		cachedResponse, err := app.redis_client.Get(context.TODO(), "defaultValues").Result()
		if err != nil {
			switch {
			case (err.Error() == "redis: nil"):
				fmt.Println("empty cache")
			default:
				logger.LogError("cache error", err)
			}
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(cachedResponse))
			if err != nil {
				utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
			}
			return
		}
	}

	input := struct {
		Title string
		Tags  []string
		Owner *int
		Flag  *int
		data.Filters
	}{}

	ownership := utils.ReadStringParam(qs, "owner", "all")
	if ownership == "me" {
		user := app.contextGetUser(r)
		input.Owner = &user.User_id
	} else if ownership == "-me" {
		user := app.contextGetUser(r)
		input.Flag = &user.User_id
	}

	input.Title = utils.ReadStringParam(qs, "title", "")
	input.Tags = utils.ReadCSVParam(qs, "tags", []string{})
	input.Filters.Page = utils.ReadIntParam(qs, "page", 1)
	input.Filters.PageSize = utils.ReadIntParam(qs, "page_size", 20)
	input.Filters.Sort = utils.ReadStringParam(qs, "sort", "document_id")
	input.Filters.SortSafelist = []string{"document_id", "-document_id"}

	documents, metadata, err := app.data_access.Documents.GetAll(input.Title, input.Tags, input.Owner, input.Flag, input.Filters)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	responses_slice := []interface{}{}

	for _, document := range documents {
		doc := struct {
			ID          int       `json:"document_id"`
			User_id     int       `json:"user_id"`
			Title       string    `json:"title"`
			Link        string    `json:"link"`
			Tags        []string  `json:"tags"`
			Uploaded_at time.Time `json:"created_at"`
		}{
			ID:          document.Document_id,
			User_id:     document.User_id,
			Title:       document.Title,
			Link:        document.Url_s3,
			Tags:        document.Tags,
			Uploaded_at: document.Uploaded_at,
		}

		responses_slice = append(responses_slice, doc)
	}

	jsonData, err := utils.WriteJSONCache(w, http.StatusOK, utils.Wrap{"metadata": metadata, "documents": responses_slice}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}

	err = app.redis_client.Set(context.TODO(), "defaultValues", jsonData, time.Hour*24).Err()
	if err != nil {
		logger.LogError("failed caching response", err)
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
		User_id:   user.User_id,
		Filetype:  ".pdf",
		Title:     file_data.Filename,
		Tags:      input.Tags,
		Is_hidden: input.Is_hidden,
	}

	disposition := "inline"
	contentType := "application/pdf"

	uploader := manager.NewUploader(app.s3_client)
	res, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:             aws.String(os.Getenv("AWS_S3_BUCKET_NAME")),
		Key:                aws.String(document.Title),
		Body:               file,
		ACL:                "public-read",
		ContentDisposition: &disposition,
		ContentType:        &contentType,
	})
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	document.Url_s3 = res.Location

	err = app.data_access.Documents.Insert(document)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	headers := http.Header{}
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", document.Document_id))

	err = utils.WriteJSON(w, http.StatusCreated, utils.Wrap{"document": document}, headers)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}

	err = app.redis_client.FlushAll(context.TODO()).Err()
	if err != nil {
		logger.LogError("Failed flushing cache", err)
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
		utils.NotFoundResponse(w, r) //? http.StatusNotFound - 404
		return
	}

	document, err := app.data_access.Documents.Get(id)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	user := app.contextGetUser(r)

	if document.User_id != user.User_id && !user.Is_admin {
		utils.InvalidCredentialsResponse(w, r)
		return
	}

	output, err := app.s3_client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String("viadro-api"),
		Key:    aws.String(document.Title),
	})
	fmt.Println("OUTPUT: ", output)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	err = app.data_access.Documents.Delete(id)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"message": "document successfully deleted"}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}

	err = app.redis_client.FlushAll(context.TODO()).Err()
	if err != nil {
		logger.LogError("Failed flushing cache", err)
	}
}

// Get document details
//
//	@Summary      Get document details
//	@Description  Get document details
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
		utils.NotFoundResponse(w, r) //? http.StatusNotFound - 404
		return
	}

	document, err := app.data_access.Documents.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			utils.NotFoundResponse(w, r) //? http.StatusNotFound - 404
		default:
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		}
		return
	}

	user := app.contextGetUser(r)

	if document.Is_hidden && document.User_id != user.User_id && !user.Is_admin {
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
		utils.NotFoundResponse(w, r) //? http.StatusNotFound - 404
		return
	}

	document, err := app.data_access.Documents.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			utils.NotFoundResponse(w, r) //? http.StatusNotFound - 404
		default:
			utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		}
		return
	}

	user := app.contextGetUser(r)

	if document.User_id != user.User_id && !user.Is_admin {
		utils.InvalidCredentialsResponse(w, r) //? http.StatusUnauthorized - 401
		return
	}

	document, err = app.data_access.Documents.ToggleVisibility(id)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
		return
	}

	response := struct {
		ID          int       `json:"document_id"`
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

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"document": response}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err) //? http.StatusInternalServerError - 500
	}

	err = app.redis_client.FlushAll(context.TODO()).Err()
	if err != nil {
		logger.LogError("Failed flushing cache", err)
	}
}
