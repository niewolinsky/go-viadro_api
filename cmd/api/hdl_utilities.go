package main

import (
	"net/http"

	_ "viadro_api/docs"
	"viadro_api/internal/logger"
	"viadro_api/utils"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	err := utils.WriteJSON(w, http.StatusOK, utils.Wrap{"status": "Status OK"}, nil)
	if err != nil {
		logger.LogError("Unable to send healthcheckHandler response", err)
	}
}

func (app *application) documentationHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	httpSwagger.WrapHandler(res, req)
}
