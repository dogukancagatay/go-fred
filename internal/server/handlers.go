package server

import (
	"net/http"

	"go-fred-rest/internal/models"

	"github.com/gin-gonic/gin"
)

// healthCheck returns the health status of the server
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "go-fred-rest",
	})
}

// createTask creates a new task
func (s *Server) createTask(c *gin.Context) {
	var req models.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := s.taskManager.CreateTask(req.Type, req.Input, req.Async)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := models.TaskResponse{Task: task}
	c.JSON(http.StatusCreated, response)
}

// listTasks returns all tasks
func (s *Server) listTasks(c *gin.Context) {
	tasks := s.taskManager.ListTasks()

	response := models.TaskListResponse{
		Tasks: make([]models.Task, len(tasks)),
		Total: len(tasks),
	}

	for i, task := range tasks {
		response.Tasks[i] = *task
	}

	c.JSON(http.StatusOK, response)
}

// getTask returns a specific task by ID
func (s *Server) getTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := models.TaskResponse{Task: task}
	c.JSON(http.StatusOK, response)
}

// executeTask executes a task synchronously
func (s *Server) executeTask(c *gin.Context) {
	taskID := c.Param("id")

	err := s.taskManager.ExecuteTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the updated task
	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := models.TaskResponse{Task: task}
	c.JSON(http.StatusOK, response)
}

// executeTaskAsync executes a task asynchronously
func (s *Server) executeTaskAsync(c *gin.Context) {
	taskID := c.Param("id")

	err := s.taskManager.ExecuteTaskAsync(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the task to return current status
	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := models.TaskResponse{Task: task}
	c.JSON(http.StatusAccepted, response)
}

// cancelTask cancels a running task
func (s *Server) cancelTask(c *gin.Context) {
	taskID := c.Param("id")

	err := s.taskManager.CancelTask(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the updated task
	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := models.TaskResponse{Task: task}
	c.JSON(http.StatusOK, response)
}

// getTaskTypes returns all supported task types
func (s *Server) getTaskTypes(c *gin.Context) {
	// Get the registry from task manager (we need to expose this method)
	// For now, we'll return the known types
	types := []string{"echo", "sleep", "error", "math"}

	c.JSON(http.StatusOK, gin.H{
		"task_types": types,
	})
}
