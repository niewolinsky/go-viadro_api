package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.Handle(http.MethodGet, "/v1/documentation/:any", app.documentationHandler)

	router.HandlerFunc(http.MethodPost, "/v1/documents", app.addDocumentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/documents", app.listDocumentsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/documents/:id", app.getDocumentHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/documents/:id", app.deleteDocumentHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/documents/:id", app.toggleDocumentVisibilityHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users/register", app.userRegister)
	router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.userActivate)
	router.HandlerFunc(http.MethodPost, "/v1/users/login", app.userLogin)
	router.HandlerFunc(http.MethodPost, "/v1/users/logout", app.userLogout)

	router.HandlerFunc(http.MethodPost, "/v1/awstest", app.s3Test)

	return router
}
