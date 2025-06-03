package service

import (
	"errors"
	"fileserver/config"
	"fileserver/internal/models"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetDocument retrieves a document from the database based on its `idFile` field.
// It searches for a document with the given `idFile` and returns the document if found, or an error if not.
func GetDocument(idFile uuid.UUID) (*models.Document, error) {
	var document models.Document

	// Perform the query to find the document by its unique `idFile` field
	if err := config.DB.Where("id_file = ?", idFile).First(&document).Error; err != nil {
		// If no record is found, return a descriptive error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("document with idFile %v not found", idFile)
		}
		// If there is another error during retrieval, return the error
		return nil, fmt.Errorf("error while retrieving document: %v", err)
	}

	// Return the document found in the database
	return &document, nil
}

// AddDocument adds a new document to the database.
// The function receives a pointer to a `Document` struct and attempts to insert it into the database.
func AddDocument(document *models.Document) error {
	// Create a new record for the document in the database
	if err := config.DB.Create(document).Error; err != nil {
		// If an error occurs during the insert, return the error
		return fmt.Errorf("error while adding document: %v", err)
	}
	// If the operation is successful, return nil (no error)
	return nil
}
