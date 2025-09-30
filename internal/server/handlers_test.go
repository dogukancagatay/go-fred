package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-fred-rest/internal/config"
	"go-fred-rest/internal/events"
	"go-fred-rest/internal/models"
	"go-fred-rest/internal/tasks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPublisher is a mock event publisher for testing
type mockPublisher struct {
	events []events.Event
}

func (m *mockPublisher) Publish(ctx context.Context, event events.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

func (m *mockPublisher) GetEvents() []events.Event {
	return m.events
}

func (m *mockPublisher) ClearEvents() {
	m.events = nil
}

func setupTestServer() *Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Events: config.EventsConfig{
			Publisher: "noop",
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 10,
		},
	}

	// Create event publisher
	eventPub := &mockPublisher{}

	// Create task executor registry and register default executors
	registry := tasks.NewExecutorRegistry()
	tasks.RegisterDefaultExecutors(registry)

	// Create task manager
	taskManager := tasks.NewTaskManager(registry, eventPub, cfg.Tasks.MaxConcurrent)

	// Create Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{
		config:      cfg,
		router:      router,
		taskManager: taskManager,
		eventPub:    eventPub,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

func TestHealthCheck(t *testing.T) {
	server := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "go-fred-rest", response["service"])
}

func TestCreateTask(t *testing.T) {
	server := setupTestServer()

	taskRequest := models.TaskRequest{
		Type:  "echo",
		Input: map[string]interface{}{"message": "hello"},
		Async: false,
	}

	jsonData, _ := json.Marshal(taskRequest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Task)
	assert.Equal(t, "echo", response.Task.Type)
	assert.Equal(t, models.TaskStatusPending, response.Task.Status)
	assert.Equal(t, "hello", response.Task.Input["message"])
	assert.False(t, response.Task.IsAsync)
}

func TestCreateTaskInvalidType(t *testing.T) {
	server := setupTestServer()

	taskRequest := models.TaskRequest{
		Type:  "invalid",
		Input: map[string]interface{}{},
		Async: false,
	}

	jsonData, _ := json.Marshal(taskRequest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestCreateTaskInvalidJSON(t *testing.T) {
	server := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListTasks(t *testing.T) {
	server := setupTestServer()

	// Create some tasks
	task1, err := server.taskManager.CreateTask("echo", map[string]interface{}{"message": "hello1"}, false)
	require.NoError(t, err)

	task2, err := server.taskManager.CreateTask("sleep", map[string]interface{}{"duration": 5}, true)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TaskListResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Tasks, 2)

	// Check that both tasks are in the response
	found1, found2 := false, false
	for _, task := range response.Tasks {
		if task.ID == task1.ID {
			found1 = true
		}
		if task.ID == task2.ID {
			found2 = true
		}
	}

	assert.True(t, found1, "Task 1 not found in response")
	assert.True(t, found2, "Task 2 not found in response")
}

func TestGetTask(t *testing.T) {
	server := setupTestServer()

	// Create a task
	task, err := server.taskManager.CreateTask("echo", map[string]interface{}{"message": "hello"}, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks/"+task.ID, nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Task)
	assert.Equal(t, task.ID, response.Task.ID)
	assert.Equal(t, "echo", response.Task.Type)
}

func TestGetTaskNotFound(t *testing.T) {
	server := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks/non-existent", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestExecuteTask(t *testing.T) {
	server := setupTestServer()

	// Create a task
	task, err := server.taskManager.CreateTask("echo", map[string]interface{}{"message": "hello"}, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks/"+task.ID+"/execute", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Task)
	assert.Equal(t, models.TaskStatusCompleted, response.Task.Status)
	assert.NotNil(t, response.Task.Output)
}

func TestExecuteTaskNotFound(t *testing.T) {
	server := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks/non-existent/execute", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestExecuteTaskAsync(t *testing.T) {
	server := setupTestServer()

	// Create a task
	task, err := server.taskManager.CreateTask("sleep", map[string]interface{}{"duration": 0.1}, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks/"+task.ID+"/execute-async", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response models.TaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Task)
	assert.Equal(t, models.TaskStatusRunning, response.Task.Status)

	// Wait for task to complete
	time.Sleep(200 * time.Millisecond)

	// Check that task completed
	completedTask, err := server.taskManager.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TaskStatusCompleted, completedTask.Status)
}

func TestExecuteTaskAsyncNotFound(t *testing.T) {
	server := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks/non-existent/execute-async", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestCancelTask(t *testing.T) {
	server := setupTestServer()

	// Create a task
	task, err := server.taskManager.CreateTask("echo", map[string]interface{}{}, false)
	require.NoError(t, err)

	// Start the task
	task.Start()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/tasks/"+task.ID, nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Task)
	assert.Equal(t, models.TaskStatusCancelled, response.Task.Status)
}

func TestCancelTaskNotFound(t *testing.T) {
	server := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/tasks/non-existent", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestGetTaskTypes(t *testing.T) {
	server := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/task-types", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	taskTypes, ok := response["task_types"].([]interface{})
	require.True(t, ok)

	expectedTypes := []string{"echo", "sleep", "error", "math"}
	assert.Len(t, taskTypes, len(expectedTypes))

	for _, expectedType := range expectedTypes {
		found := false
		for _, taskType := range taskTypes {
			if taskType == expectedType {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected task type %s not found", expectedType)
	}
}

func TestCORS(t *testing.T) {
	server := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/tasks", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
}
