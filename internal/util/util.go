package util

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func WriteBadRequest(w http.ResponseWriter, err error, description string, log *logrus.Logger) {
	writeErrorResponse(w, err, description, log, http.StatusBadRequest)
}

func WriteInternalServer(w http.ResponseWriter, err error, description string, log *logrus.Logger) {
	writeErrorResponse(w, err, description, log, http.StatusInternalServerError)
}

func writeErrorResponse(w http.ResponseWriter, err error, description string, log *logrus.Logger, statusCode int) {
	if description != "" {
		log.WithError(err).Errorln(description)
	} else {
		log.Errorln(err)
	}
	w.WriteHeader(statusCode)
}
