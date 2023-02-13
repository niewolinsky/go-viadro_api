package main

import (
	"encoding/json"
	"net/http"

	_ "viadro_api/docs"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	res, _ := json.MarshalIndent("Status OK", "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func (app *application) documentationHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	httpSwagger.WrapHandler(res, req)
}
