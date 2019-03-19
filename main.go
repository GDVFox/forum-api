package main

import (
	"forum-api/controllers"
	"forum-api/models"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Handler структура хэндлера запросов
type Handler struct {
	Router *mux.Router
}

func main() {
	connectError := models.ConnetctDB("forum_db_user", "qwerty", "localhost", "forum_db")
	if connectError != nil {
		log.Fatalf("cant open database connection: %s", connectError.Message)
	}

	r := mux.NewRouter().PathPrefix("/api/v1").Subrouter()
	r.HandleFunc("/user/{nickname}/profile", controllers.GetUser).Methods("GET")
	r.HandleFunc("/user/{nickname}/create", controllers.CreateUser).Methods("POST")
	r.HandleFunc("/user/{nickname}/profile", controllers.UpdateUser).Methods("POST")

	r.HandleFunc("/forum/create", controllers.CreateForum).Methods("POST")
	r.HandleFunc("/forum/{slug}/details", controllers.GetForum).Methods("GET")

	r.HandleFunc("/forum/{slug}/create", controllers.CreateThread).Methods("POST")
	r.HandleFunc("/forum/{slug}/threads", controllers.GetThreadsByForum).Methods("GET")

	h := Handler{
		Router: r,
	}

	port := os.Getenv("PORT")
	log.Printf("MainService successfully started at port %s", port)
	err := http.ListenAndServe(":"+port, h.Router)
	if err != nil {
		log.Fatalf("cant start main server. err: %s", err.Error())
	}
}
