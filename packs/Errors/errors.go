package Errors

import (
	// "forumDb/models"
	"net/http"
	"github.com/kirillsonk/forumDb/models"
)

func CheckDuplicateError(name string) string {
	res := "duplicate key value violates unique constraint \"" + name + "\""
	return res
}

func SendError(errText string, statusCode int, w *http.ResponseWriter) ([]byte, error) {
	e := new(models.Error)
	e.Message = errText
	// resData, _ := e.MarshalJSON()
	resData, _ := e.MarshalJSON()


	(*w).Header().Set("content-type", "application/json")
	(*w).WriteHeader(statusCode)
	(*w).Write(resData)

	return resData, nil
}
