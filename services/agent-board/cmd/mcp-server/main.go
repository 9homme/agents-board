package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent-board/internal/config"
	"agent-board/internal/fsutil"
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
	// Resolve DATABASE_URL via the config helper (D-006: DB_URL rejected at startup).
	dbURL, err := config.ResolveDBURL()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("db config: using DATABASE_URL")

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Signal-cancellable lifecycle context (D-008): SIGINT + SIGTERM cancel the boot-time ping.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Bounded DB ping: 5-second timeout derived from lifecycle context (D-013).
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second) // TODO(REQ005): make configurable if ops needs it
	defer cancel()

	if err := pingDB(pingCtx, db); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}

	projectRepo := repo.NewProjectRepo(db)
	documentRepo := repo.NewDocumentRepo(db)
	userStoryRepo := repo.NewUserStoryRepo(db)
	taskRepo := repo.NewTaskRepo(db)
	requirementRepo := repo.NewRequirementRepo(db)

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

	validator := fsutil.NewFsValidator()
	handler.RegisterProjectTools(toolRegistry, projectRepo, validator)
	handler.RegisterDocumentTools(toolRegistry, documentRepo, requirementRepo)
	handler.RegisterUserStoryTools(toolRegistry, userStoryRepo, requirementRepo)
	handler.RegisterTaskTools(toolRegistry, taskRepo)
	handler.RegisterAuditTools(toolRegistry, repo.NewAuditRepo(db))
	handler.RegisterRequirementTools(toolRegistry, requirementRepo)

	h := handler.NewHandler(sessionManager, toolRegistry)

	e.GET("/sse", h.HandleSSE)
	e.POST("/message", h.HandleMessage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return e.Start(":" + port)
}

// pingDB calls db.PingContext with the provided context.
// The caller is responsible for constructing a context with an appropriate
// deadline and/or signal cancellation before calling this function.
func pingDB(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}
