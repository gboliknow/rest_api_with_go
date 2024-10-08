package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr  string
	store Store
}

func NewAPIServer(addr string, store Store) *APIServer {
	return &APIServer{addr: addr, store: store}
}

func (s *APIServer) Serve() {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	//registering the routes

	userService := NewUserService(s.store)
	userService.RegisterRoutes(subrouter)

	projectService := NewProjectService(s.store)
	projectService.RegisterRoutes(subrouter)

	tasksService := NewTasksService(s.store)
	tasksService.RegisterRoutes(subrouter)
	log.Fatal(http.ListenAndServe(s.addr, subrouter))
}
