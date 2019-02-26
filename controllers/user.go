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
		var code int
		if err.Code == models.InternalDatabase {
			code = http.StatusInternalServerError
			err.Message = "internal database error"
		} else if err.Code == models.RowNotFound {
			code = http.StatusNotFound
			err.Message = "Can't find user with that nickname"
		}

		utils.WriteEasyjson(w, code, err)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, userInfo)
}

// CreateUser создание нового пользователя в базе данных.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	newUser := &models.User{}
	err := utils.DecodeEasyjson(r.Body, newUser)
	if err != nil {
		utils.WriteEasyjson(w, http.StatusBadRequest, &models.Error{
			Message: "unable to decode request body;",
		})
		return
	}

	newUser.Nickname = vars["nickname"]
	used, errs := newUser.Create()
	if used != nil {
		utils.WriteEasyjson(w, http.StatusConflict, used)
		return
	}

	if errs != nil {
		var code int
		if errs.Code == models.ValidationFailed {
			code = http.StatusBadRequest
			errs.Message = "Validation faild"
		} else {
			code = http.StatusInternalServerError
			errs.Message = "internal server error"
		}

		utils.WriteEasyjson(w, code, errs)
		return
	}

	utils.WriteEasyjson(w, http.StatusCreated, newUser)
}
