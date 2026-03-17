// app/models/post.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Post represents a blog post.
//
// BUG FIX: The original imported "gorm.io/datatypes" for JSONMap but that
// package is a separate optional module not listed in go.mod — importing it
// causes "cannot find module providing gorm.io/datatypes" at go mod tidy.
// Replaced with a plain string (store JSON as text) which works with all drivers,
// or use map[string]interface{} with a custom Scanner/Valuer if needed.
type Post struct {
	ID        uint           `gorm:"primaryKey"                           json:"id"`
	Title     string         `gorm:"type:varchar(255);not null"           json:"title"`
	Slug      string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Content   string         `gorm:"type:text"                            json:"content"`
	Published bool           `gorm:"default:false"                        json:"published"`
	Views     int            `gorm:"default:0"                            json:"views"`
	UserID    uint           `gorm:"index"                                json:"user_id"`
	MetaJSON  string         `gorm:"type:text;default:'{}'"               json:"meta"` // store JSON as string
	CreatedAt time.Time      `                                            json:"created_at"`
	UpdatedAt time.Time      `                                            json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                json:"-"`

	// Relationship — loaded via Preload("User")
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName overrides the default table name.
func (Post) TableName() string { return "posts" }

// BeforeSave validates the post before writing to the database.
func (p *Post) BeforeSave(tx *gorm.DB) error {
	if len(p.Title) < 5 {
		return gorm.ErrInvalidData
	}
	return nil
}

// AfterCreate fires after a new post row is inserted.
func (p *Post) AfterCreate(tx *gorm.DB) error {
	// Emit event, update search index, etc.
	return nil
}

// AfterUpdate fires after an existing post row is updated.
func (p *Post) AfterUpdate(tx *gorm.DB) error {
	// Invalidate cache, etc.
	return nil
}

// IsPublished returns true when the post is visible to readers.
func (p *Post) IsPublished() bool { return p.Published }

// IncrementViews atomically increments the view counter.
func (p *Post) IncrementViews(db *gorm.DB) error {
	return db.Model(p).Update("views", gorm.Expr("views + ?", 1)).Error
}

// ─────────────────────────────────────────────────────────────────────────────

// User represents an application user.
type User struct {
	ID        uint           `gorm:"primaryKey"                             json:"id"`
	Name      string         `gorm:"type:varchar(255);not null"             json:"name"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null"             json:"-"` // never serialised
	Role      string         `gorm:"type:varchar(50);default:'user'"        json:"role"`
	CreatedAt time.Time      `                                              json:"created_at"`
	UpdatedAt time.Time      `                                              json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                  json:"-"`

	// Relationship
	Posts []Post `gorm:"foreignKey:UserID" json:"posts,omitempty"`
}

// TableName overrides the default table name.
func (User) TableName() string { return "users" }

// BeforeCreate hashes the password before the first insert.
// Replace the stub with a real bcrypt call in production.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password == "" {
		return gorm.ErrInvalidData
	}
	// TODO: u.Password = bcrypt.Hash(u.Password)
	return nil
}
