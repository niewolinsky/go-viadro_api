package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Wrap map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Wrap, headers http.Header) error {
	json, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	json = append(json, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(json)

	return nil
}

func WriteJSONCache(w http.ResponseWriter, status int, data Wrap, headers http.Header) ([]byte, error) {
	json, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return nil, err
	}

	json = append(json, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(json)

	return json, nil
}

func ReadJSON(w http.ResponseWriter, r *http.Request, source interface{}) error {
	requestLimit := 5242880
	r.Body = http.MaxBytesReader(w, r.Body, int64(requestLimit))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(source)
	if err != nil {
		return err
	}

	return nil
}

func ReadMultipartJSON(w http.ResponseWriter, r *http.Request, source interface{}) (multipart.File, *multipart.FileHeader, error) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, err
	}

	metadata := r.FormValue("metadata")
	b := bytes.NewBufferString(metadata)

	decoder := json.NewDecoder(b)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(source)
	if err != nil {
		return nil, nil, err
	}

	file, file_data, err := r.FormFile("document")
	if err != nil {
		return nil, nil, err
	}

	return file, file_data, nil
}

func ReadIDParam(r *http.Request) (int, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return int(id), nil
}

func ReadStringParam(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

func ReadCSVParam(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func ReadIntParam(qs url.Values, key string, defaultValue int) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return i
}

func CacheSave() {

}

func CacheRead() {

}
