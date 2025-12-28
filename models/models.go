package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID        string         `gorm:"primaryKey;type:varchar(25)" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Role      Role           `gorm:"type:varchar(20);default:'user'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	News   []News  `gorm:"foreignKey:AuthorID" json:"news,omitempty"`
	Events []Event `gorm:"foreignKey:AuthorID" json:"events,omitempty"`
}

// BeforeCreate hook to generate CUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = cuid.New()
	}
	return nil
}

type News struct {
	ID         string         `gorm:"primaryKey;type:varchar(25)" json:"id"`
	Title      string         `gorm:"not null" json:"title"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Excerpt    string         `gorm:"type:varchar(200)" json:"excerpt"` // Plain text excerpt
	Slug       string         `gorm:"uniqueIndex;not null" json:"slug"`
	Published  bool           `gorm:"default:false" json:"published"`
	ImageURL   string         `json:"image_url"`
	AuthorID   string         `gorm:"type:varchar(25);not null" json:"author_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Author User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// BeforeCreate hook to generate CUID
func (n *News) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = cuid.New()
	}
	return nil
}

type Event struct {
	ID         string         `gorm:"primaryKey;type:varchar(25)" json:"id"`
	Title      string         `gorm:"not null" json:"title"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Excerpt    string         `gorm:"type:varchar(200)" json:"excerpt"` // Plain text excerpt
	Slug       string         `gorm:"uniqueIndex;not null" json:"slug"`
	Published  bool           `gorm:"default:false" json:"published"`
	ImageURL   string         `json:"image_url"`
	AuthorID   string         `gorm:"type:varchar(25);not null" json:"author_id"`
	EventStart *time.Time     `json:"event_start"` // Start date & time of event
	EventEnd   *time.Time     `json:"event_end"`   // End date & time of event
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Author User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// BeforeCreate hook to generate CUID
func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = cuid.New()
	}
	return nil
}

type Hole struct {
	ID          string         `gorm:"primaryKey;type:varchar(25)" json:"id"`
	HoleIndex   int            `gorm:"default:0" json:"hole_index"` // Order/Sequence number
	Name        string         `gorm:"not null" json:"name"` // e.g., "Hole 1"
	Description string         `gorm:"type:text" json:"description"`
	Par         int            `gorm:"default:0" json:"par"`      // Par value for this hole
	Distance    int            `gorm:"default:0" json:"distance"` // Distance in meters
	ImageURL    string         `json:"image_url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate CUID
func (h *Hole) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = cuid.New()
	}
	return nil
}


