package Errors

import (
	"ForumsApi/models"
	"encoding/json"
	"net/http"
)

func SendError(errText string, statusCode int, w *http.ResponseWriter) ([]byte, error){
	e := new(models.Error)
	e.Message = errText
	resp, _ := json.Marshal(e)

	// Проверка err json

	(*w).Header().Set("content-type", "application/json")
	(*w).WriteHeader(statusCode)
	(*w).Write(resp)

	return resp, nil
}
