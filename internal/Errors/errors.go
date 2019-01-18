package Errors

import (
	// "forumDb/models"
	"net/http"

	"bitbucket.org/kirillsonk/forumDb/models"
)

func SendError(errText string, statusCode int, w *http.ResponseWriter) ([]byte, error) {
	e := new(models.Error)
	e.Message = errText
	resData, _ := e.MarshalJSON()

	(*w).Header().Set("content-type", "application/json")
	(*w).WriteHeader(statusCode)
	(*w).Write(resData)

	return resData, nil
}
