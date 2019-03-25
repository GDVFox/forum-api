package controllers

import (
	"net/http"

	"github.com/GDVFox/forum-api/models"
	"github.com/GDVFox/forum-api/utils"

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
		} else if err.Code == models.RowNotFound {
			code = http.StatusNotFound
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
		} else {
			code = http.StatusInternalServerError
		}

		utils.WriteEasyjson(w, code, errs)
		return
	}

	utils.WriteEasyjson(w, http.StatusCreated, newUser)
}

// UpdateUser изменение информации в профиле пользователя.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	updateFields := &models.UpdateUserFields{}
	jsonErr := utils.DecodeEasyjson(r.Body, updateFields)
	if jsonErr != nil {
		utils.WriteEasyjson(w, http.StatusBadRequest, &models.Error{
			Message: "unable to decode request body;",
		})
		return
	}

	user, err := models.GetUserByNickname(vars["nickname"])
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

	if updateFields.Fullname != nil {
		user.Fullname = *updateFields.Fullname
	}
	if updateFields.About != nil {
		user.About = *updateFields.About
	}
	if updateFields.Email != nil {
		user.Email = *updateFields.Email
	}

	err = user.Save()
	if err != nil {
		var code int
		if err.Code == models.InternalDatabase {
			code = http.StatusInternalServerError
		} else if err.Code == models.RowDuplication {
			code = http.StatusConflict
		}

		utils.WriteEasyjson(w, code, err)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, user)
}
