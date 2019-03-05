package controllers

import (
	"forum-api/models"
	"forum-api/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// GetForum получение информации о форуме
func GetForum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	forumInfo, err := models.GetForumBySlug(vars["slug"])
	if err != nil {
		var code int
		if err.Code == models.InternalDatabase {
			code = http.StatusInternalServerError
		} else if err.Code == models.RowNotFound {
			code = http.StatusNotFound
		}

		utils.WriteEasyjson(w, code, err)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, forumInfo)
}

// CreateForum создание нового пользователя в базе данных.
func CreateForum(w http.ResponseWriter, r *http.Request) {
	newForum := &models.Forum{}
	err := utils.DecodeEasyjson(r.Body, newForum)
	if err != nil {
		utils.WriteEasyjson(w, http.StatusBadRequest, &models.Error{
			Message: "unable to decode request body;",
		})
		return
	}

	used, errs := newForum.Create()
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

	utils.WriteEasyjson(w, http.StatusCreated, newForum)
}
