package main

import (
	"fmt"
	"net/http"

	"viadro_api/internal/data"
	"viadro_api/utils"
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
		fmt.Println(err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"documents": documents}, nil)
	if err != nil {
		fmt.Println("Server error 2")
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
		fmt.Println("Bad response")
		return
	}

	fmt.Println(input)

	document := &data.Document{
		Url_s3:     input.Url_s3,
		Filetype:   input.Filetype,
		Title:      input.Title,
		Tags:       input.Tags,
		Is_private: input.Is_private,
	}

	fmt.Println(document)

	err = app.data_access.Documents.Insert(document)
	if err != nil {
		fmt.Println(err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", document.Document_id))

	fmt.Println(document)
	err = utils.WriteJSON(w, http.StatusCreated, utils.Wrap{"document": document}, headers)
	if err != nil {
		fmt.Println("Server error")
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
	fmt.Println(id)
	if err != nil {
		fmt.Println("Not found")
		return
	}

	err = app.data_access.Documents.Delete(id)
	if err != nil {
		fmt.Println("Bad request")
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"message": "document successfully deleted"}, nil)
	if err != nil {
		fmt.Println("Server error")
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
	fmt.Println(id)
	if err != nil {
		fmt.Println("Not found")
		return
	}

	document, err := app.data_access.Documents.Get(id)
	if err != nil {
		fmt.Println("Bad request")
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Wrap{"document": document}, nil)
	if err != nil {
		fmt.Println("Server error 2")
	}
}
