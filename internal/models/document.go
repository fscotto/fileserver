package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// Document represents the structure of the documents table in the database.
type Document struct {
	ID          uint           `gorm:"primaryKey"`                      // Primary key for the document
	Name        string         `gorm:"column:name"`                     // Name of the document
	IdFile      uuid.UUID      `gorm:"type:uuid;column:id_file;unique"` // Unique identifier for the document's file
	Fingerprint string         `gorm:"column:fingerprint;unique"`       // Unique fingerprint (hash) for the document
	CreatedAt   time.Time      `gorm:"column:created_at"`               // Timestamp of when the document was created
	UpdatedAt   time.Time      `gorm:"column:updated_at"`               // Timestamp of when the document was last updated
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at"`         // Timestamp for soft deletion (if applicable)
}

// TableName overrides the default table name used by GORM.
func (Document) TableName() string {
	// Returns the name of the table where documents are stored
	return "documents"
}
