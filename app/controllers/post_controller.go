// app/controllers/post_controller.go
package controllers

import (
	"github.com/gofiber/fiber/v2"
	veloHttp "github.com/manish-npx/velo/http"
	"github.com/manish-npx/velo/validation"
)

// PostController handles post-related requests.
// BUG FIX: All methods now use veloHttp.Handler signature (func(*veloHttp.Context) error).
// The original used func(*veloHttp.Context) error on the method but the router
// accepted fiber.Handler — those are incompatible types and cause compile errors.
type PostController struct {
	// postService *services.PostService
}

func NewPostController() *PostController {
	return &PostController{}
}

// ── Request structs ──────────────────────────────────────────────────────────

type CreatePostRequest struct {
	Title   string `json:"title"   validate:"required,min=5,max=255"`
	Content string `json:"content" validate:"required,min=10"`
	Slug    string `json:"slug"    validate:"required,min=3"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"   validate:"omitempty,min=5,max=255"`
	Content string `json:"content" validate:"omitempty,min=10"`
}

// ── Handlers — all are veloHttp.Handler (func(*veloHttp.Context) error) ──────

// Index  GET /posts
func (c *PostController) Index(ctx *veloHttp.Context) error {
	posts := []map[string]interface{}{
		{"id": 1, "title": "Getting Started with Velo", "slug": "getting-started"},
	}
	return ctx.JSON(fiber.Map{"data": posts, "total": len(posts)})
}

// Show  GET /posts/:id
func (c *PostController) Show(ctx *veloHttp.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(fiber.Map{
		"id":    id,
		"title": "Getting Started with Velo",
	})
}

// Store  POST /posts
func (c *PostController) Store(ctx *veloHttp.Context) error {
	var req CreatePostRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	v := validation.NewValidator()
	if err := v.Validate(req); err != nil {
		return ctx.Status(422).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(201).JSON(fiber.Map{
		"message": "Post created successfully",
		"data":    fiber.Map{"id": 1, "title": req.Title, "slug": req.Slug},
	})
}

// Update  PUT /posts/:id
func (c *PostController) Update(ctx *veloHttp.Context) error {
	id := ctx.Param("id")

	var req UpdatePostRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	v := validation.NewValidator()
	if err := v.Validate(req); err != nil {
		return ctx.Status(422).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{
		"message": "Post updated successfully",
		"data":    fiber.Map{"id": id, "title": req.Title},
	})
}

// Delete  DELETE /posts/:id
// BUG FIX: Original declared `id := ctx.Param("id")` but never used it —
// Go treats unused variables as compile errors. Fixed by using id in the response.
func (c *PostController) Delete(ctx *veloHttp.Context) error {
	id := ctx.Param("id")
	return ctx.Status(200).JSON(fiber.Map{
		"message": "Post deleted successfully",
		"id":      id,
	})
}

// Search  GET /posts?q=keyword&limit=10
func (c *PostController) Search(ctx *veloHttp.Context) error {
	query := ctx.Query("q")
	limit := ctx.Query("limit", "10") // BUG FIX: original called ctx.Query("limit", "10")
	//                                   but the old signature was Query(key string) string
	//                                   with no default support — would compile but silently
	//                                   ignore the default. Now fixed in Context.Query.
	return ctx.JSON(fiber.Map{
		"query": query,
		"limit": limit,
		"data":  []interface{}{},
	})
}

// ── Implement ResourceController interface ────────────────────────────────────
// So PostController can be passed to router.Resource("/posts", ctrl)

var _ veloHttp.ResourceController = (*PostController)(nil) // compile-time check

// (Index, Show, Store, Update, Delete already defined above — interface satisfied)
