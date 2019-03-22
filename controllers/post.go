package controllers

import (
	"forum-api/models"
	"forum-api/utils"
	"net/http"
	"strconv"
	"time"

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

	creationTime := time.Now()
	for _, p := range newPosts {
		p.Created = creationTime
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

func GetPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	slug := vars["slug_or_id"]
	threadID, err := strconv.ParseInt(slug, 10, 64)
	isID := (err == nil)

	query := r.URL.Query()
	limitParam, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		limitParam = -1
	}
	offsetParam, _ := strconv.ParseInt(query.Get("since"), 10, 64)
	desc := (query.Get("desc") == "true")

	mode := models.Flat
	switch query.Get("sort") {
	case "flat":
		mode = models.Flat
	case "tree":
		mode = models.Tree
	case "parent_tree":
		mode = models.ParentTree
	}

	var posts models.Posts
	var getError *models.Error
	if isID {
		posts, getError = models.GetPostsByThreadID(threadID, limitParam, offsetParam, mode, desc)
	} else {
		posts, getError = models.GetPostsByThreadSlug(slug, limitParam, offsetParam, mode, desc)
	}
	if getError != nil {
		if getError.Code == models.RowNotFound {
			utils.WriteEasyjson(w, http.StatusNotFound, getError)
			return
		}

		utils.WriteEasyjson(w, http.StatusInternalServerError, getError)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, posts)
}
