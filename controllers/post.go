package controllers

import (
	"forum-api/models"
	"forum-api/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func CreatePosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	slug := vars["slug_or_id"]
	threadID, err := strconv.ParseInt(slug, 10, 64)
	isID := (err == nil)

	newPosts := make(models.Posts, 0)
	err = utils.DecodeEasyjson(r.Body, &newPosts)
	if err != nil {
		utils.WriteEasyjson(w, http.StatusBadRequest, &models.Error{
			Message: "unable to decode request body;",
		})
		return
	}

	for _, p := range newPosts {
		if isID {
			p.Thread = threadID
		} else {
			p.ThreadSlug = &slug
		}
	}

	if createError := newPosts.Create(); createError != nil {
		var code int
		if createError.Code == models.ValidationFailed {
			code = http.StatusBadRequest
		} else if createError.Code == models.ForeignKeyNotFound {
			code = http.StatusNotFound
		} else {
			code = http.StatusInternalServerError
		}

		utils.WriteEasyjson(w, code, createError)
		return
	}

	utils.WriteEasyjson(w, http.StatusCreated, newPosts)
}
