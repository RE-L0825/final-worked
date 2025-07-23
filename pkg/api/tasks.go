package api

import (
	"net/http"
	"strconv"

	"github.com/RE-L0825/final-worked/pkg/db"
)

type TaskResponse struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := db.GetTasks(50)
	if err != nil {
		sendError(w, "Failed to get tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := TasksResponse{
		Tasks: make([]TaskResponse, 0, len(tasks)),
	}

	for _, task := range tasks {
		response.Tasks = append(response.Tasks, TaskResponse{
			ID:      strconv.FormatInt(task.ID, 10),
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		})
	}

	sendJSON(w, response, http.StatusOK)
}
