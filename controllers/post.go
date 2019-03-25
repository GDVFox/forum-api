package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GDVFox/forum-api/models"
	"github.com/GDVFox/forum-api/utils"

	"github.com/gorilla/mux"
)

func CreatePosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var slugOrID interface{}
	slug := vars["slug_or_id"]
	slugOrID = slug

	threadID, err := strconv.ParseInt(slug, 10, 64)
	if err == nil {
		slugOrID = threadID
	}

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
	}

	if createError := newPosts.Create(slugOrID); createError != nil {
		var code int
		if createError.Code == models.ValidationFailed {
			code = http.StatusBadRequest
		} else if createError.Code == models.ForeignKeyNotFound {
			code = http.StatusNotFound
		} else if createError.Code == models.ForeignKeyConflict {
			code = http.StatusConflict
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

func GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]
	postID, _ := strconv.ParseInt(id, 10, 64)

	info, err := models.GetPostByID(postID, strings.Split(r.URL.Query().Get("related"), ","))
	if err != nil {
		if err.Code == models.RowNotFound {
			utils.WriteEasyjson(w, http.StatusNotFound, err)
			return
		}

		utils.WriteEasyjson(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, info)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]
	postID, _ := strconv.ParseInt(id, 10, 64)
	post := &models.Post{}
	err := utils.DecodeEasyjson(r.Body, post)
	if err != nil {
		utils.WriteEasyjson(w, http.StatusBadRequest, &models.Error{
			Message: "unable to decode request body;",
		})
		return
	}
	post.ID = postID

	if updErr := post.Update(); updErr != nil {
		if updErr.Code == models.RowNotFound {
			utils.WriteEasyjson(w, http.StatusNotFound, updErr)
			return
		}

		utils.WriteEasyjson(w, http.StatusInternalServerError, updErr)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, post)
}
