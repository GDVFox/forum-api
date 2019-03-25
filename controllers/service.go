package controllers

import (
	"net/http"

	"github.com/GDVFox/forum-api/models"
	"github.com/GDVFox/forum-api/utils"
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
	err := models.Load()
	if err != nil {
		utils.WriteEasyjson(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
