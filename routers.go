package main

import (
	"database/sql"

	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/tasks", handleCreateTask(db)).Methods("POST")
	r.HandleFunc("/tasks", handleGetTasks(db)).Methods("GET")
	r.HandleFunc("/tasks/{id:[0-9]+}", handleGetTask(db)).Methods("GET")
	r.HandleFunc("/tasks/{id:[0-9]+}", handleUpdateTask(db)).Methods("PUT")
	r.HandleFunc("/tasks/{id:[0-9]+}", handleDeleteTask(db)).Methods("DELETE")
	r.HandleFunc("/tags", handleCreateTag(db)).Methods("POST")
	r.HandleFunc("/tags", handleGetTags(db)).Methods("GET")
	r.HandleFunc("/stats", handleGetStats(db)).Methods("GET")
	return r
}
