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
		fmt.Println("Server error 1")
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
		Title string `json:"title"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		fmt.Println("Bad response")
		return
	}

	document := &data.Document{
		Title: input.Title,
	}

	err = app.data_access.Documents.Insert(document)
	if err != nil {
		fmt.Println("Server error")
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
