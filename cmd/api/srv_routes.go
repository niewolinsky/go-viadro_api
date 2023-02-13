package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/documents", app.addDocumentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/documents", app.listDocumentsHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/documents/:id", app.deleteDocumentHandler)

	return router
}
