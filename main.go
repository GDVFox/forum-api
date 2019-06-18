package main

import (
	"log"
	"net/http"

	"github.com/GDVFox/forum-api/controllers"
	"github.com/GDVFox/forum-api/models"

	"github.com/gorilla/mux"
)

// Handler структура хэндлера запросов
type Handler struct {
	Router *mux.Router
}

func main() {
	//connectError := models.ConnetctDB("docker", "docker", "localhost", "docker")
	connectError := models.ConnetctDB("forum_db_user", "qwerty", "localhost", "forum_db")
	if connectError != nil {
		log.Fatalf("cant open database connection: %s", connectError.Message)
	}

	r := mux.NewRouter().PathPrefix("/api").Subrouter()
	r.HandleFunc("/user/{nickname}/profile", controllers.GetUser).Methods("GET")
	r.HandleFunc("/user/{nickname}/create", controllers.CreateUser).Methods("POST")
	r.HandleFunc("/user/{nickname}/profile", controllers.UpdateUser).Methods("POST")

	r.HandleFunc("/forum/create", controllers.CreateForum).Methods("POST")
	r.HandleFunc("/forum/{slug}/details", controllers.GetForum).Methods("GET")
	r.HandleFunc("/forum/{slug}/create", controllers.CreateThread).Methods("POST")
	r.HandleFunc("/forum/{slug}/threads", controllers.GetThreadsByForum).Methods("GET")
	r.HandleFunc("/forum/{slug}/users", controllers.GetForumUsers).Methods("GET")

	r.HandleFunc("/thread/{slug_or_id}/create", controllers.CreatePosts).Methods("POST")
	r.HandleFunc("/thread/{slug_or_id}/vote", controllers.Vote).Methods("POST")
	r.HandleFunc("/thread/{slug_or_id}/details", controllers.GetThread).Methods("GET")
	r.HandleFunc("/thread/{slug_or_id}/posts", controllers.GetPosts).Methods("GET")
	r.HandleFunc("/thread/{slug_or_id}/details", controllers.UpdateThread).Methods("POST")

	r.HandleFunc("/post/{id:[0-9]+}/details", controllers.GetPost).Methods("GET")
	r.HandleFunc("/post/{id:[0-9]+}/details", controllers.UpdatePost).Methods("POST")

	r.HandleFunc("/service/status", controllers.GetStatus).Methods("GET")
	r.HandleFunc("/service/clear", controllers.Clear).Methods("POST")

	h := Handler{
		Router: r,
	}

	port := "5000"
	log.Printf("MainService successfully started at port %s", port)
	err := http.ListenAndServe(":"+port, h.Router)
	if err != nil {
		log.Fatalf("cant start main server. err: %s", err.Error())
	}
}
