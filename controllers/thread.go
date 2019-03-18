package controllers

import (
	"forum-api/models"
	"forum-api/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// CreateForum создание нового пользователя в базе данных.
func CreateThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	newThread := &models.Thread{}
	err := utils.DecodeEasyjson(r.Body, newThread)
	if err != nil {
		utils.WriteEasyjson(w, http.StatusBadRequest, &models.Error{
			Message: "unable to decode request body;",
		})
		return
	}

	newThread.Forum = vars["slug"]
	used, errs := newThread.Create()
	if used != nil {
		utils.WriteEasyjson(w, http.StatusConflict, used)
		return
	}

	if errs != nil {
		var code int
		if errs.Code == models.ValidationFailed {
			code = http.StatusBadRequest
		} else if errs.Code == models.ForeignKeyNotFound {
			code = http.StatusNotFound
		} else {
			code = http.StatusInternalServerError
		}

		utils.WriteEasyjson(w, code, errs)
		return
	}

	utils.WriteEasyjson(w, http.StatusCreated, newThread)
}
