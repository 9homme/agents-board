package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"agent-board/internal/handler"
	"agent-board/internal/repo"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Printf("api-server exited with error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Configure CORS
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "*"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontendURL},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Read DATABASE_URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Signal-cancellable lifecycle context (D-008): SIGINT + SIGTERM cancel the boot-time ping.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Bounded DB ping: 5-second timeout derived from lifecycle context (D-013).
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pingDB(pingCtx, db); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}

	projectRepo := repo.NewProjectRepo(db)
	projectHandler := handler.NewProjectHandler(projectRepo)

	documentHandler := handler.NewDocumentHandler(repo.NewDocumentRepo(db), projectRepo)

	e.GET("/api/v1/projects", projectHandler.GetProjects)
	e.GET("/api/v1/projects/:id", projectHandler.GetProject)
	e.GET("/api/v1/projects/:id/documents", documentHandler.ListProjectDocuments)
	e.GET("/api/v1/documents/:id", documentHandler.GetDocument)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Strip control characters from port before logging to prevent log injection (G706).
	// PORT is an env var expected to contain only digits; sanitising is a defence-in-depth measure.
	safePort := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, port)
	// G706: safePort has all control chars (< 0x20 and DEL) stripped above; log injection is not possible.
	log.Printf("Starting api-server on port %s", safePort) //nolint:gosec
	return e.Start(":" + port)
}

// pingDB calls db.PingContext with the provided context.
// The caller is responsible for constructing a context with an appropriate
// deadline and/or signal cancellation before calling this function.
func pingDB(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}
