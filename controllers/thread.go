package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/GDVFox/forum-api/models"
	"github.com/GDVFox/forum-api/utils"

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

func UpdateThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	slug := vars["slug_or_id"]
	threadID, err := strconv.ParseInt(slug, 10, 64)
	isID := (err == nil)

	thread := &models.Thread{}
	err = utils.DecodeEasyjson(r.Body, thread)
	if err != nil {
		utils.WriteEasyjson(w, http.StatusBadRequest, &models.Error{
			Message: "unable to decode request body;",
		})
		return
	}

	if isID {
		thread.ID = threadID
	} else {
		thread.Slug = &slug
	}

	if updErr := thread.Update(); updErr != nil {
		if updErr.Code == models.RowNotFound {
			utils.WriteEasyjson(w, http.StatusNotFound, updErr)
			return
		}

		utils.WriteEasyjson(w, http.StatusInternalServerError, updErr)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, thread)
}

func GetThreadsByForum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	query := r.URL.Query()
	limitParam, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		limitParam = -1
	}
	offsetParam, _ := time.Parse(time.RFC3339Nano, query.Get("since"))
	desc := (query.Get("desc") == "true")

	threads, threadsErr := models.GetThreadsByForum(vars["slug"], limitParam, offsetParam, desc)
	if threadsErr != nil {
		if threadsErr.Code == models.RowNotFound {
			utils.WriteEasyjson(w, http.StatusNotFound, threadsErr)
			return
		}

		utils.WriteEasyjson(w, http.StatusInternalServerError, threadsErr)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, threads)
}

func Vote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	slug := vars["slug_or_id"]
	threadID, err := strconv.ParseInt(slug, 10, 64)
	isID := (err == nil)

	voice := &models.Vote{}
	err = utils.DecodeEasyjson(r.Body, voice)
	if err != nil {
		utils.WriteEasyjson(w, http.StatusBadRequest, &models.Error{
			Message: "unable to decode request body;",
		})
		return
	}

	var thread *models.Thread
	var threadErr *models.Error
	if isID {
		thread, threadErr = models.VoteByID(threadID, voice)
	} else {
		thread, threadErr = models.VoteBySlug(slug, voice)
	}
	if threadErr != nil {
		if threadErr.Code == models.RowNotFound || threadErr.Code == models.ForeignKeyNotFound {
			utils.WriteEasyjson(w, http.StatusNotFound, threadErr)
			return
		}

		utils.WriteEasyjson(w, http.StatusInternalServerError, threadErr)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, thread)
}

func GetThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	slug := vars["slug_or_id"]
	threadID, err := strconv.ParseInt(slug, 10, 64)
	isID := (err == nil)

	var thread *models.Thread
	var threadErr *models.Error
	if isID {
		thread, threadErr = models.GetThreadByID(threadID)
	} else {
		thread, threadErr = models.GetThreadBySlug(slug)
	}
	if threadErr != nil {
		if threadErr.Code == models.RowNotFound {
			utils.WriteEasyjson(w, http.StatusNotFound, threadErr)
			return
		}

		utils.WriteEasyjson(w, http.StatusInternalServerError, threadErr)
		return
	}

	utils.WriteEasyjson(w, http.StatusOK, thread)
}
