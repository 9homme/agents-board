package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"agent-board/internal/handler"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Printf("mcp-server exited with error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := db.PingContext(context.Background()); err != nil {
		return err
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
	handler.RegisterAuditTools(toolRegistry, repo.NewAuditRepo(db))

	h := handler.NewHandler(sessionManager, toolRegistry)

	e.GET("/sse", h.HandleSSE)
	e.POST("/message", h.HandleMessage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return e.Start(":" + port)
}
