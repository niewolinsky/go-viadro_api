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
	router.HandlerFunc(http.MethodGet, "/v1/documents", app.listAllVisibleDocumentsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/documents/my", app.requireActivatedUser(app.listUserDocumentsHandler))

	//?single document routes
	router.HandlerFunc(http.MethodGet, "/v1/document/:id", app.getDocumentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/document", app.requireActivatedUser(app.addDocumentHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/document/:id", app.requireActivatedUser(app.deleteDocumentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/document/:id", app.requireActivatedUser(app.toggleDocumentVisibilityHandler))

	//?user authentication routes
	router.HandlerFunc(http.MethodPost, "/v1/user/register", app.userRegister)
	router.HandlerFunc(http.MethodPut, "/v1/user/activate", app.userActivate)
	router.HandlerFunc(http.MethodPut, "/v1/user/authenticate", app.userAuthenticate)
	//! deleting account should result in all documents being deleted from bucket
	router.HandlerFunc(http.MethodDelete, "/v1/user/:id", app.requireActivatedUser(app.userDelete))

	//?admin routes + can delete/get all documents and users
	router.HandlerFunc(http.MethodPatch, "/v1/admin/user/:id", app.requireAdminUser(app.toggleAdminGrant))
	router.HandlerFunc(http.MethodGet, "/v1/documents/all", app.requireAdminUser(app.listAllDocumentsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/admin/users", app.requireAdminUser(app.getAllUsers))

	return app.authenticate(router)
}
