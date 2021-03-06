package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	InternalServerErrMessage = "Internal server error"
)

type StatusCode int

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func WriteBadRequest(w http.ResponseWriter, r *http.Request, userMessage string) {
	log.WithFields(log.Fields{
		"method":   r.Method,
		"url":      r.URL,
		"httpCode": http.StatusBadRequest,
	}).Warn(userMessage)

	WriteError(w, r, userMessage, http.StatusBadRequest)
}

func WriteServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.WithFields(log.Fields{
		"method":   r.Method,
		"url":      r.URL,
		"httpCode": http.StatusInternalServerError,
	}).Errorf("\n%+v\n", err)

	WriteError(w, r, InternalServerErrMessage, http.StatusInternalServerError)
}

func WriteError(w http.ResponseWriter, r *http.Request, userMessage string, httpErrorCode StatusCode) {
	w.WriteHeader(int(httpErrorCode))
	WriteJSON(w, r, &ErrorMessage{int(httpErrorCode), userMessage})
}

func WriteJSON(w http.ResponseWriter, r *http.Request, payload interface{}) {
	w.Header().Set("content-type", "application/json")

	jsonResponse, err := json.Marshal(payload)
	if err != nil {
		WriteServerError(w, r, err)
		return
	}

	WriteResponse(w, r, jsonResponse)
}

func WriteResponse(w io.Writer, r *http.Request, resBytes []byte) {
	bytesWritten, err := w.Write(resBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"method":   r.Method,
			"url":      r.URL,
			"resBytes": fmt.Sprintf("[% x]", resBytes),
		}).Errorf("Err writing response bytes. \n%+v\n", err)

		return
	}

	log.WithFields(log.Fields{
		"method":       r.Method,
		"url":          r.URL,
		"bytesWritten": bytesWritten,
		"httpCode":     http.StatusOK,
	}).Info("HTTP response sent.")
}
