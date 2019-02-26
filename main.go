package main

import (
	"forum-api/controllers"
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
	r := mux.NewRouter()
	r.HandleFunc("/user/{nickname}/profile", controllers.GetUser).Methods("GET")
	r.HandleFunc("/user/{nickname}/create", controllers.CreateUser).Methods("POST")

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
