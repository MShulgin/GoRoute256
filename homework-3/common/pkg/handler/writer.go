package handler

import (
	"encoding/json"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"net/http"
)

func WriteJson(httpStatus int, w http.ResponseWriter, payload any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("cannot write response: " + err.Error())
	}
}

func WriteError(w http.ResponseWriter, err *ex.AppError) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(err.HttpCode())
	if err := json.NewEncoder(w).Encode(err); err != nil {
		logger.Error("cannot encode app error: " + err.Error())
	}
}
