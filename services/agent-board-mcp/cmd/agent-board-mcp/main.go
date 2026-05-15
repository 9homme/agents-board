package main

import (
	"database/sql"
	"log"
	"os"

	"agent-board-mcp/internal/handler"
	"agent-board-mcp/internal/mcp"
	"agent-board-mcp/internal/repo"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	projectRepo := repo.NewProjectRepo(db)
	documentRepo := repo.NewDocumentRepo(db)
	userStoryRepo := repo.NewUserStoryRepo(db)
	taskRepo := repo.NewTaskRepo(db)

	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Printf("REQUEST: method: %v, uri: %v, status: %v\n", v.Method, v.URI, v.Status)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	sessionManager := mcp.NewSessionManager()
	toolRegistry := mcp.NewToolRegistry()

	handler.RegisterProjectTools(toolRegistry, projectRepo)
	handler.RegisterDocumentTools(toolRegistry, documentRepo)
	handler.RegisterUserStoryTools(toolRegistry, userStoryRepo)
	handler.RegisterTaskTools(toolRegistry, taskRepo)

	h := handler.NewHandler(sessionManager, toolRegistry)

	e.GET("/sse", h.HandleSSE)
	e.POST("/message", h.HandleMessage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(e.Start(":" + port))
}
