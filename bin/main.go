// bin/main.go — complete working example using the fixed Velo API
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manish-npx/velo/config"
	"github.com/manish-npx/velo/core"
	"github.com/manish-npx/velo/database"
	veloHttp "github.com/manish-npx/velo/http"
	"github.com/manish-npx/velo/logger"
	"github.com/manish-npx/velo/validation"
)

// ── Request structs ──────────────────────────────────────────────────────────

type CreateUserRequest struct {
	Name     string `json:"name"     validate:"required,min=3,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// ── Controller ───────────────────────────────────────────────────────────────

type UserController struct{}

// All handler methods are func(*veloHttp.Context) error — i.e. veloHttp.Handler.
// BUG FIX DEMONSTRATION: the original code passed these to router.Get() which
// expected fiber.Handler (func(*fiber.Ctx) error). That caused a compile error
// on every route registration. The fix is in router.go: Get/Post/etc. now
// accept veloHttp.Handler directly and adapt internally.

func (c *UserController) Index(ctx *veloHttp.Context) error {
	return ctx.JSON(fiber.Map{
		"data":    []string{},
		"message": "list users",
	})
}

func (c *UserController) Show(ctx *veloHttp.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(fiber.Map{"id": id})
}

func (c *UserController) Store(ctx *veloHttp.Context) error {
	var req CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
	}
	v := validation.NewValidator()
	if err := v.Validate(req); err != nil {
		return ctx.Status(422).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.Status(201).JSON(fiber.Map{
		"message": "user created",
		"name":    req.Name,
		"email":   req.Email,
	})
}

func (c *UserController) Update(ctx *veloHttp.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(fiber.Map{"id": id, "message": "updated"})
}

// BUG FIX: Delete now uses the id param in the response — avoids
// "id declared but not used" compile error from the original.
func (c *UserController) Delete(ctx *veloHttp.Context) error {
	id := ctx.Param("id")
	return ctx.Status(200).JSON(fiber.Map{"id": id, "message": "deleted"})
}

// ── Service Provider ─────────────────────────────────────────────────────────

type AppServiceProvider struct{}

func (p *AppServiceProvider) Register(app *core.App) error { return nil }

func (p *AppServiceProvider) Boot(app *core.App) error {
	cfg := config.NewConfig(".env")
	dbCfg := cfg.GetDatabaseConfig()

	db, err := database.Connect(database.Connection{
		Driver: dbCfg.Driver,
		Path:   cfg.GetWithDefault("DB_PATH", "./database/app.db"),
	})
	if err != nil {
		app.Logger.Warn("Database not connected (set DB_PATH or DB_DRIVER in .env)", err)
		return nil // non-fatal in dev
	}
	app.SetDB(db)
	app.Logger.Info("Database connected", dbCfg.Driver)
	return nil
}

// ── Bootstrap & routes ───────────────────────────────────────────────────────

func main() {
	cfg := config.NewConfig(".env")
	appCfg := cfg.GetAppConfig()

	log := logger.NewSimpleLogger(logger.LevelInfo)
	defer log.Close()

	app := core.New(map[string]interface{}{
		"name":       appCfg.Name,
		"env":        appCfg.Env,
		"port":       appCfg.Port,
		"jwt_secret": cfg.GetWithDefault("JWT_SECRET", "change-me"),
	}, log)

	if err := app.Register(&AppServiceProvider{}); err != nil {
		log.Error("Provider registration failed", err)
		return
	}

	router := veloHttp.NewRouter(app)

	// Global middleware
	router.Use(
		veloHttp.LoggingMiddleware(app),
		veloHttp.ErrorHandler(app),
		veloHttp.CORS([]string{"*"}),
	)

	// ── Public routes ────────────────────────────────────────────────────────
	router.Get("/", func(c *veloHttp.Context) error {
		return c.JSON(fiber.Map{
			"framework": "Velo",
			"version":   "0.1.0",
			"env":       appCfg.Env,
		})
	})

	router.Get("/health", func(c *veloHttp.Context) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ── API v1 group (requires JWT) ──────────────────────────────────────────
	jwtCfg := veloHttp.JWTConfig{Secret: cfg.GetWithDefault("JWT_SECRET", "change-me")}

	api := router.Group("/api/v1", veloHttp.JWTAuth(jwtCfg))

	// RESTful resource — all 5 routes registered via ResourceController interface
	ctrl := &UserController{}
	api.Resource("/users", ctrl)

	// Nested route example — BUG FIX: Group now returns *Router so you can
	// chain Velo handlers; the original returned fiber.Router which required
	// the raw Fiber handler signature.
	admin := api.Group("/admin", veloHttp.AdminMiddleware)
	admin.Get("/stats", func(c *veloHttp.Context) error {
		return c.JSON(fiber.Map{"users": 0, "posts": 0})
	})

	// ── Query default value demo ─────────────────────────────────────────────
	// BUG FIX: ctx.Query("limit", "10") was silently broken — the old signature
	// was Query(key string) string with no default param. Now works correctly.
	router.Get("/search", func(c *veloHttp.Context) error {
		q := c.Query("q")
		limit := c.Query("limit", "10") // default "10" when absent
		return c.JSON(fiber.Map{"query": q, "limit": limit})
	})

	if err := app.Start(appCfg.Port); err != nil {
		log.Error("Server failed", err)
	}
}
