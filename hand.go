package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func handleCreateTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		if task.Title == "" {
			writeError(w, http.StatusBadRequest, "Title is required")
			return
		}
		if task.Priority == "" {
			task.Priority = "medium"
		}
		if task.Status == "" {
			task.Status = "pending"
		}
		createdTask, err := CreateTask(db, task)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, createdTask)
	}
}

func handleGetTasks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		filters := map[string]string{
			"status":   query.Get("status"),
			"priority": query.Get("priority"),
			"tag":      query.Get("tag"),
			"search":   query.Get("search"),
			"sort_by":  query.Get("sort_by"),
			"order":    query.Get("order"),
			"limit":    query.Get("limit"),
			"offset":   query.Get("offset"),
		}
		tasks, err := GetTasks(db, filters)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, tasks)
	}
}

func handleGetTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}
		task, err := GetTask(db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "Task not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, task)
	}
}

func handleUpdateTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		if task.Title == "" {
			writeError(w, http.StatusBadRequest, "Title is required")
			return
		}
		updatedTask, err := UpdateTask(db, id, task)
		if err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "Task not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, updatedTask)
	}
}

func handleDeleteTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}
		err = DeleteTask(db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "Task not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Task deleted"})
	}
}

func handleCreateTag(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		if input.Name == "" {
			writeError(w, http.StatusBadRequest, "Name is required")
			return
		}
		tag, err := CreateTag(db, input.Name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, tag)
	}
}

func handleGetTags(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tags, err := GetTags(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, tags)
	}
}

func handleGetStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := GetStats(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, stats)
	}
}
