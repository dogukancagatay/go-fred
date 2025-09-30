package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go-fred/internal/config"
	"go-fred/internal/events"
	"go-fred/internal/tasks"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	config      *config.Config
	router      *gin.Engine
	taskManager *tasks.TaskManager
	eventPub    events.Publisher
	httpServer  *http.Server
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	// Create event publisher
	eventPub, err := events.NewPublisher(&cfg.Events)
	if err != nil {
		log.Fatalf("Failed to create event publisher: %v", err)
	}

	// Create task executor registry and register default executors
	registry := tasks.NewExecutorRegistry()
	tasks.RegisterDefaultExecutors(registry)

	// Create task manager
	taskManager := tasks.NewTaskManager(registry, eventPub, cfg.Tasks.MaxConcurrent)

	// Create Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

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

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Task management endpoints
		v1.POST("/tasks", s.createTask)
		v1.GET("/tasks", s.listTasks)
		v1.GET("/tasks/:id", s.getTask)
		v1.POST("/tasks/:id/execute", s.executeTask)
		v1.POST("/tasks/:id/execute-async", s.executeTaskAsync)
		v1.DELETE("/tasks/:id", s.cancelTask)

		// Task types endpoint
		v1.GET("/task-types", s.getTaskTypes)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	address := s.config.GetAddress()

	s.httpServer = &http.Server{
		Addr:    address,
		Handler: s.router,
	}

	log.Printf("Starting server on %s", address)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	// Close event publisher
	if err := s.eventPub.Close(); err != nil {
		log.Printf("Error closing event publisher: %v", err)
	}

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
