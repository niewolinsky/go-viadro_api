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

	//?document routes
	router.HandlerFunc(http.MethodGet, "/v1/documents", app.getAllDocumentsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/document/:id", app.getDocumentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/document", app.requireActivatedUser(app.addDocumentHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/document/:id", app.requireActivatedUser(app.deleteDocumentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/document/:id", app.requireActivatedUser(app.toggleDocumentVisibilityHandler))

	//?user routes
	router.HandlerFunc(http.MethodPost, "/v1/user", app.userRegisterHandler)
	router.HandlerFunc(http.MethodPut, "/v1/user/activate", app.userActivateHandler)
	router.HandlerFunc(http.MethodPut, "/v1/user/authenticate", app.userAuthenticateHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/user/:id", app.requireActivatedUser(app.userDeleteHandler))

	//?admin routes
	router.HandlerFunc(http.MethodPatch, "/v1/admin/user/:id", app.requireAdminUser(app.toggleAdminGrantHandler))
	router.HandlerFunc(http.MethodGet, "/v1/admin/users", app.requireAdminUser(app.getAllUsersHandler))
	router.HandlerFunc(http.MethodGet, "/v1/admin/documents", app.requireAdminUser(app.getAllDocumentsAdminHandler))

	return app.authenticate(router)
}
