package controllers

import (
	"forum-api/models"
	"forum-api/utils"
	"net/http"
)

func GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := models.GetStatus()
	if err != nil {
		utils.WriteEasyjson(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, status)
}

func Clear(w http.ResponseWriter, r *http.Request) {
	err := models.Clear()
	if err != nil {
		utils.WriteEasyjson(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
