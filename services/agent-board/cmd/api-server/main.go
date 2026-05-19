package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"

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

	if err := db.PingContext(context.Background()); err != nil {
		return err
	}

	projectRepo := repo.NewProjectRepo(db)
	projectHandler := handler.NewProjectHandler(projectRepo)

	e.GET("/api/v1/projects", projectHandler.GetProjects)

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
