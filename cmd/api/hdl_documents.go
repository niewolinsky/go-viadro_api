package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"viadro_api/internal/data"
	"viadro_api/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// listDocumentsHandler godoc
// @Summary      get all public documents
// @Description  get all public documents
// @Tags         documents
// @Accept       json
// @Produce      json
// @Router       /documents [get]
func (app *application) listAllDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	documents, err := app.data_access.Documents.GetAll()
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
		}{
			ID:          document.Document_id,
			Title:       document.Title,
			Link:        document.Url_s3,
			Tags:        document.Tags,
			Uploaded_at: document.Uploaded_at,
		}

		responseSlice = append(responseSlice, doc)
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"documents": responseSlice}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

func (app *application) listUserDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	documents, err := app.data_access.Documents.GetUserAll(user.ID)
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

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"documents": responseSlice}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

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

func (app *application) getDocumentHandler(w http.ResponseWriter, r *http.Request) {
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

	if document.Is_hidden && document.User_id != user.ID {
		utils.InvalidCredentialsResponse(w, r)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"document": document}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

func (app *application) toggleDocumentVisibilityHandler(w http.ResponseWriter, r *http.Request) {
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

func (app *application) s3Test(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // maxMemory 32MB
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	file, _, err := r.FormFile("document")
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}
	defer file.Close()

	//!bugged, not opening in browser but downloading as attachment
	uploader := manager.NewUploader(app.s3_client)
	response, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("viadro-api"),
		Key:    aws.String("sample_pdf.pdf"),
		Body:   file,
		ACL:    "public-read",
	})
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}
	fmt.Println(response.Location)

	w.WriteHeader(200)
}
