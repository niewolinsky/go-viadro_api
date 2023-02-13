package main

import (
	"net/http"

	_ "viadro_api/docs"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *application) documentationHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	httpSwagger.WrapHandler(res, req)
}

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/documents", app.addDocumentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/documents", app.listDocumentsHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/documents/:id", app.deleteDocumentHandler)
	router.Handle(http.MethodGet, "/v1/documentation/:any", app.documentationHandler)

	return router
}
