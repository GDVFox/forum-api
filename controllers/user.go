package controllers

import (
	"forum-api/models"
	"forum-api/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// GetUser получение информации о пользователе форума по его имени.
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userInfo, err := models.GetUserByNickname(vars["nickname"])
	if err != nil {
		if err.Code == models.InternalDatabase {
			w.WriteHeader(http.StatusInternalServerError)
			err.Message = "internal database error"
		} else if err.Code == models.RowNotFound {
			w.WriteHeader(http.StatusNotFound)
			err.Message = "Can't find user with that nickname"
		}

		utils.WriteEasyjson(err, w)
		return
	}

	utils.WriteEasyjson(userInfo, w)
}
