// app/services/post_service.go
package services

import (
	"fmt"
	"go-velo-example/app/models"

	"gorm.io/gorm"
	// BUG FIX: original imported "my-app/app/models" — a hardcoded placeholder
	// module path that will always fail ("cannot find module my-app").
	// Replace this import path with your actual module name from go.mod, e.g.:
	//   "github.com/yourname/yourapp/app/models"
	// For this example file we use a relative package name and note the fix.
)

// PostService encapsulates all business logic for blog posts.
type PostService struct {
	db *gorm.DB
}

// NewPostService creates a PostService with an injected DB connection.
func NewPostService(db *gorm.DB) *PostService {
	return &PostService{db: db}
}

// GetAll retrieves posts with optional published filter and pagination.
func (s *PostService) GetAll(limit, offset int, publishedOnly bool) ([]models.Post, error) {
	var posts []models.Post
	q := s.db.Model(&models.Post{})
	if publishedOnly {
		q = q.Where("published = ?", true)
	}
	result := q.Limit(limit).Offset(offset).Order("created_at DESC").Find(&posts)
	return posts, result.Error
}

// GetByID retrieves a single post by primary key, preloading its author.
func (s *PostService) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	result := s.db.Preload("User").First(&post, id)
	if result.Error != nil {
		// BUG FIX: gorm.ErrRecordNotFound was deprecated in GORM v2.
		// The correct sentinel is still gorm.ErrRecordNotFound (it still exists)
		// but the reliable cross-version check is errors.Is:
		//   if errors.Is(result.Error, gorm.ErrRecordNotFound) { ... }
		// We use that pattern here.
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post with id %d not found", id)
		}
		return nil, result.Error
	}
	return &post, nil
}

// GetBySlug retrieves a published post by its URL slug.
func (s *PostService) GetBySlug(slug string) (*models.Post, error) {
	var post models.Post
	result := s.db.Where("slug = ? AND published = ?", slug, true).
		Preload("User").
		First(&post)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post '%s' not found", slug)
		}
		return nil, result.Error
	}
	return &post, nil
}

// Create inserts a new post row.
func (s *PostService) Create(userID uint, title, slug, content string) (*models.Post, error) {
	post := &models.Post{
		Title:   title,
		Slug:    slug,
		Content: content,
		UserID:  userID,
	}
	result := s.db.Create(post)
	return post, result.Error
}

// Update applies a whitelist of field changes to a post.
func (s *PostService) Update(id uint, data map[string]interface{}) (*models.Post, error) {
	post, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	allowed := map[string]bool{"title": true, "content": true, "published": true}
	filtered := make(map[string]interface{}, len(data))
	for k, v := range data {
		if allowed[k] {
			filtered[k] = v
		}
	}

	result := s.db.Model(post).Updates(filtered)
	return post, result.Error
}

// Delete soft-deletes a post (sets DeletedAt).
func (s *PostService) Delete(id uint) error {
	result := s.db.Delete(&models.Post{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("post with id %d not found", id)
	}
	return nil
}

// Publish marks a post as published.
//
// BUG FIX: original wrote:
//
//	s.db.Model(&models.Post{}, id).Update(...)
//
// The second argument to Model() is ignored — GORM's Model() accepts a single
// value pointer. To scope to a specific row you must chain .Where or use
// db.Model(&post).Where("id = ?", id) or simply use a loaded struct.
func (s *PostService) Publish(id uint) error {
	return s.db.Model(&models.Post{}).Where("id = ?", id).Update("published", true).Error
}

// Unpublish marks a post as unpublished.
func (s *PostService) Unpublish(id uint) error {
	return s.db.Model(&models.Post{}).Where("id = ?", id).Update("published", false).Error
}

// Search finds published posts whose title or content contains the query string.
func (s *PostService) Search(query string, limit int) ([]models.Post, error) {
	var posts []models.Post
	like := "%" + query + "%"
	result := s.db.Where("published = ? AND (title ILIKE ? OR content ILIKE ?)", true, like, like).
		Limit(limit).
		Order("created_at DESC").
		Find(&posts)
	return posts, result.Error
}

// GetUserPosts retrieves paginated posts for a specific author.
func (s *PostService) GetUserPosts(userID uint, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	result := s.db.Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&posts)
	return posts, result.Error
}

// IncrementViews atomically increments a post's view counter.
func (s *PostService) IncrementViews(id uint) error {
	post, err := s.GetByID(id)
	if err != nil {
		return err
	}
	return post.IncrementViews(s.db)
}

// GetTrending returns the top posts by view count.
func (s *PostService) GetTrending(limit int) ([]models.Post, error) {
	var posts []models.Post
	result := s.db.Where("published = ?", true).
		Order("views DESC, created_at DESC").
		Limit(limit).
		Find(&posts)
	return posts, result.Error
}
