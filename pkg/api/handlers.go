package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/RE-L0825/final-worked/pkg/db"

)

func sendError(w http.ResponseWriter, message string, statusCode int) {
	sendJSON(w, map[string]string{"error": message}, statusCode)
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode ...int) {
	w.Header().Set("Content-Type", "application/json")
	if len(statusCode) > 0 {
		w.WriteHeader(statusCode[0])
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to send JSON response: %v", err)
	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendError(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	sendJSON(w, map[string]string{
		"id":      strconv.FormatInt(task.ID, 10),
		"date":    task.Date,
		"title":   task.Title,
		"comment": task.Comment,
		"repeat":  task.Repeat,
	}, http.StatusOK)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		sendError(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var requestData struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	if err := json.Unmarshal(body, &requestData); err != nil {
		sendError(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(requestData.Title) == "" {
		sendError(w, "Task title is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(requestData.ID, 10, 64)
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if requestData.Date != "" {
		if _, err := time.Parse("20060102", requestData.Date); err != nil {
			sendError(w, "Invalid date format (use YYYYMMDD)", http.StatusBadRequest)
			return
		}
	}

	if requestData.Repeat != "" {
		now := time.Now()
		if _, err := NextDate(now, requestData.Date, requestData.Repeat); err != nil {
			sendError(w, "Invalid repeat rule: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	task := &db.Task{
		ID:      id,
		Date:    requestData.Date,
		Title:   requestData.Title,
		Comment: requestData.Comment,
		Repeat:  requestData.Repeat,
	}

	if err := db.UpdateTask(task); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, struct{}{}, http.StatusOK)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendError(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, struct{}{}, http.StatusOK)
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    body, err := io.ReadAll(r.Body)
    if err != nil {
        sendError(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    var task db.Task
    if err := json.Unmarshal(body, &task); err != nil {
        sendError(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
        return
    }

    if strings.TrimSpace(task.Title) == "" {
        sendError(w, "Task title is required", http.StatusBadRequest)
        return
    }

    now := time.Now()
    currentDate := now.Format("20060102")

    if task.Date == "" {
        task.Date = currentDate
    } else {
        taskTime, err := time.Parse("20060102", task.Date)
        if err != nil {
            sendError(w, "Invalid date format (use YYYYMMDD)", http.StatusBadRequest)
            return
        }

        if taskTime.Before(now.Truncate(24 * time.Hour)) {
            if task.Repeat == "" {
                task.Date = currentDate
            } else {
                nextDate, err := NextDate(now, task.Date, task.Repeat)
                if err != nil {
                    sendError(w, "Invalid repeat rule: "+err.Error(), http.StatusBadRequest)
                    return
                }
                task.Date = nextDate
            }
        }
    }

    if task.Repeat != "" {
        if _, err := NextDate(now, task.Date, task.Repeat); err != nil {
            sendError(w, "Invalid repeat rule: "+err.Error(), http.StatusBadRequest)
            return
        }
    }

    id, err := db.AddTask(&task)
    if err != nil {
        log.Printf("Database error: %v", err)
        sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    sendJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)}, http.StatusCreated)
}

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendError(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	if task.Repeat != "" {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendError(w, "Failed to calculate next date: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.UpdateTaskDate(id, nextDate); err != nil {
			sendError(w, "Failed to update task: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := db.DeleteTask(id); err != nil {
			sendError(w, "Failed to delete task: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	sendJSON(w, struct{}{}, http.StatusOK)
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse("20060102", nowStr)
		if err != nil {
			sendError(w, "invalid now date", http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
