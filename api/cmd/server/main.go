package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/phathdt/login-oauth/api/internal/auth"
	"github.com/phathdt/login-oauth/api/internal/config"
	"github.com/phathdt/login-oauth/api/internal/database"
	dbpkg "github.com/phathdt/login-oauth/api/internal/db"
	"github.com/phathdt/login-oauth/api/internal/handlers"
)

func main() {
	cfg := config.Load()

	// pgxpool is used only for migrations (goose requires *sql.DB via stdlib).
	pool, err := database.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// sql.DB via pgx stdlib for sqlc-generated queries.
	sqlDB, err := database.NewSQLDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open sql.DB: %v", err)
	}
	defer sqlDB.Close()

	queries := dbpkg.New(sqlDB)

	oauthConfig := auth.NewOAuthConfig(cfg)

	oauthHandler := handlers.NewOAuthHandler(cfg, oauthConfig, queries)
	authHandler := handlers.NewAuthHandler(cfg, queries)
	productHandler := handlers.NewProductHandler()

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization",
		AllowCredentials: true,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/auth/google/login", oauthHandler.Login)
	app.Get("/auth/google/callback", oauthHandler.Callback)
	app.Post("/auth/refresh", authHandler.Refresh)
	app.Post("/auth/logout", authHandler.Logout)

	protected := app.Group("/", auth.JWTAuth(cfg))
	protected.Get("/auth/me", authHandler.Me)
	protected.Get("/api/products", productHandler.List)

	addr := ":" + cfg.Port
	log.Printf("server starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
