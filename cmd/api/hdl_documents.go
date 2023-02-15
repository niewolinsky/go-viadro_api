package main

import (
	"context"
	"fmt"
	"net/http"

	"viadro_api/internal/data"
	"viadro_api/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// listDocumentsHandler godoc
// @Summary      get all public documents
// @Description  get all public documents
// @Tags         documents
// @Accept       json
// @Produce      json
// @Router       /documents [get]
func (app *application) listDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	documents, err := app.data_access.Documents.GetAll()
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"documents": documents}, nil)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
	}
}

// addDocumentsHandler godoc
// @Summary      add one document
// @Description  add one document
// @Tags         documents
// @Accept       json
// @Produce      json
// @Router       /documents [post]
func (app *application) addDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Url_s3    string   `json:"url_s3"`
		Filetype  string   `json:"filetype"`
		Title     string   `json:"title"`
		Tags      []string `json:"tags"`
		Is_hidden bool     `json:"is_hidden"`
	}

	file, err := utils.ReadMultipartJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	defer file.Close()

	document := &data.Document{
		Url_s3:    input.Url_s3,
		Filetype:  input.Filetype,
		Title:     input.Title,
		Tags:      input.Tags,
		Is_hidden: input.Is_hidden,
	}

	res, err := app.s3_manager.Upload(context.TODO(), &s3.PutObjectInput{
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

// deleteDocumentHandler godoc
// @Summary      delete one document
// @Description  delete one document
// @Tags         documents
// @Accept       json
// @Produce      json
// @Router       /documents/:id [delete]
func (app *application) deleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		utils.NotFoundResponse(w, r)
		return
	}

	//! no error if document does not exist
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

// listDocumentsHandler godoc
// @Summary      get details of one public documents
// @Description  get details of one public documents
// @Tags         documents
// @Accept       json
// @Produce      json
// @Router       /documents/:id [get]
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

	document, err := app.data_access.Documents.ToggleVisibility(id)
	if err != nil {
		utils.ServerErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"document": document}, nil)
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
	response, err := app.s3_manager.Upload(context.TODO(), &s3.PutObjectInput{
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
