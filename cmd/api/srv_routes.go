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
	router.HandlerFunc(http.MethodGet, "/v1/documents", app.documentGetAllHandler)
	router.HandlerFunc(http.MethodGet, "/v1/document/:id", app.documentGetHandler)
	router.HandlerFunc(http.MethodPost, "/v1/document", app.requireActivatedUser(app.documentAddHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/document/:id", app.requireActivatedUser(app.documentDeleteHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/document/:id", app.requireActivatedUser(app.documentToggleVisibilityHandler))

	//?user routes
	router.HandlerFunc(http.MethodPost, "/v1/user", app.userRegisterHandler)
	router.HandlerFunc(http.MethodPut, "/v1/user/activate", app.userActivateHandler)
	router.HandlerFunc(http.MethodPut, "/v1/user/authenticate", app.userAuthenticateHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/user/:id", app.requireActivatedUser(app.userDeleteHandler))

	//?admin routes
	router.HandlerFunc(http.MethodPatch, "/v1/admin/user/:id", app.requireAdminUser(app.adminGrantPrivilegesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/admin/users", app.requireAdminUser(app.adminGetAllUsersHandler))
	router.HandlerFunc(http.MethodGet, "/v1/admin/documents", app.requireAdminUser(app.adminGetAllDocumentsHandler))

	return app.authenticate(router)
}
