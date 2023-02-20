package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	//?utility routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.Handle(http.MethodGet, "/v1/documentation/:any", app.documentationHandler)

	//?multiple documents routes
	router.HandlerFunc(http.MethodGet, "/v1/documents", app.listAllDocumentsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/documents/my", app.listUserDocumentsHandler)

	//?single document routes
	router.HandlerFunc(http.MethodGet, "/v1/document/:id", app.getDocumentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/document", app.requireActivatedUser(app.addDocumentHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/document/:id", app.requireActivatedUser(app.deleteDocumentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/document/:id", app.requireActivatedUser(app.toggleDocumentVisibilityHandler))

	//?user authentication routes
	router.HandlerFunc(http.MethodPost, "/v1/users/register", app.userRegister)
	router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.userActivate)
	router.HandlerFunc(http.MethodPut, "/v1/users/authenticate", app.requireActivatedUser(app.userAuthenticate))

	//?dev
	router.HandlerFunc(http.MethodPost, "/v1/awstest", app.s3Test)

	return app.authenticate(router)
}
