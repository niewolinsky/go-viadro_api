package main

import (
	"net/http"

	_ "viadro_api/docs"
	"viadro_api/utils"

	"github.com/charmbracelet/log"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Check service status
//
//	@Summary      Check service status
//	@Description  Check service status
//	@Tags         utility
//	@Produce      json
//	@Success      200  {string}  "Service available"
//	@Failure      500  {string}  "Internal server error"
//	@Router       /healthcheck [get]
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	err := utils.WriteJSON(w, http.StatusOK, utils.Wrap{"status": "Status OK"}, nil)
	if err != nil {
		log.Error("Unable to send healthcheckHandler response", err)
	}
}

// API documentation
//
//	@Summary      API documentation
//	@Description  API documentation
//	@Tags         utility
//	@Produce      html
//	@Success      200  {string}  "Page loaded"
//	@Failure      404  {string}  "Page not found"
//	@Router       /documentation/index.html [get]
func (app *application) documentationHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	httpSwagger.WrapHandler(w, r)
}
