package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/stellentus/go-plc"
)

type RawTagsHandler struct {
	plc.ReadWriter
	validTags map[string]interface{}
}

func (h RawTagsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		h.Read(w, req)
	case "POST":
		h.Write(w, req)
	default:
		logHTTPError(w, http.StatusBadRequest, "Unexpected verb: "+req.Method)
	}
}

func logHTTPError(w http.ResponseWriter, code int, reason string) {
	log.Printf("Bad http request: %s (%d) - %s\n", http.StatusText(code), code, reason)
	w.WriteHeader(code)
	w.Write([]byte(reason))
}

func copy(v interface{}) interface{} {
	return reflect.New(reflect.TypeOf(v)).Interface()
}

func (h RawTagsHandler) Read(w http.ResponseWriter, req *http.Request) {
	tags := req.FormValue("tags")
	tagsToFetch := map[string]interface{}{}

	if tags != "" {
		for _, tag := range strings.Split(tags, ",") {
			v, ok := h.validTags[tag]
			if !ok {
				logHTTPError(w, http.StatusNotFound, "Tag does not exist: "+tag)
				return
			}
			tagsToFetch[tag] = copy(v)
		}
	} else {
		for tag, v := range h.validTags {
			tagsToFetch[tag] = copy(v)
		}
	}

	for tag, value := range tagsToFetch {
		h.ReadTag(tag, value)
		tagsToFetch[tag] = value
	}

	json.NewEncoder(w).Encode(tagsToFetch)
}

func (h RawTagsHandler) Write(w http.ResponseWriter, req *http.Request) {
	values := map[string]interface{}{}
	err := json.NewDecoder(req.Body).Decode(&values)

	switch {
	case err != nil:
		logHTTPError(w, http.StatusBadRequest, "failed to decode user error: "+err.Error())
		return
	case len(values) == 0:
		logHTTPError(w, http.StatusBadRequest, "empty list of tags")
		return
	default:
		for tag := range values {
			_, ok := h.validTags[tag]
			if !ok {
				logHTTPError(w, http.StatusNotFound, "Tag does not exist: "+tag)
				return
			}
		}
	}

	var lastError error
	for tag, value := range values {
		valueType := reflect.TypeOf(value)

		desiredType := reflect.TypeOf(h.validTags[tag])
		if valueType.ConvertibleTo(desiredType) {
			fetchVal := reflect.ValueOf(value).Convert(desiredType)
			err = h.WriteTag(tag, fetchVal.Interface())
		} else {
			err = fmt.Errorf("Cannot convert between types %T and %v", value, desiredType)
		}

		if err != nil {
			lastError = err
			values[tag] = err
		}
	}

	if lastError != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(values)
}
