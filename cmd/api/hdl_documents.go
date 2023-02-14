package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
		Url_s3     string   `json:"url_s3"`
		Filetype   string   `json:"filetype"`
		Title      string   `json:"title"`
		Tags       []string `json:"tags"`
		Is_private bool     `json:"is_private"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}

	document := &data.Document{
		Url_s3:     input.Url_s3,
		Filetype:   input.Filetype,
		Title:      input.Title,
		Tags:       input.Tags,
		Is_private: input.Is_private,
	}

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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Access the photo key - First Approach
	_, h, err := r.FormFile("photo")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f, openErr := os.Open(h.Filename)
	if openErr != nil {
		fmt.Println("not working")
		os.Exit(1)
	}

	response, err := app.s3_manager.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("pdfrain-sandbox-s3"),
		Key:    aws.String("cv-przemyslaw-niewolinski-en.pdf"),
		Body:   f,
		ACL:    "public-read",
	})
	if err != nil {
		log.Fatalf("failed to init uploader, %v", err)
	}
	fmt.Println(response.Location)

	w.WriteHeader(200)
}
