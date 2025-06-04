package service

import (
	"errors"
	"fileserver/config"
	"fileserver/internal/models"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetFiles retrieves a list of documents from the database based on a fuzzy search on file names.
// It only returns documents that have not been logically deleted (i.e., deleted_at is NULL).
// The function performs a case-insensitive search using the provided search query.
//
// Parameters:
//   - searchQuery (string): The search term used to find documents by their file name. This will be used
//     in a fuzzy search with the 'ILIKE' operator in PostgreSQL.
//
// Returns:
// - []models.Document: A slice of documents that match the search query and are not logically deleted.
// - error: An error is returned if there is an issue with retrieving the documents from the database.
func GetFiles(searchQuery string) ([]models.Document, error) {
	// Declare a slice to hold the results of the query
	var documents []models.Document

	// Perform the query to find documents where:
	// - 'deleted_at' is NULL (i.e., the document has not been logically deleted)
	// - The file name matches the search query using a case-insensitive pattern match ('ILIKE')
	if err := config.DB.Where("deleted_at IS NULL AND name ILIKE ?", searchQuery).Find(&documents).Error; err != nil {
		// If there is an error during the query execution, return an empty slice and the error message
		return documents, fmt.Errorf("error retrieving documents: %v", err)
	}

	// Return the list of documents and nil error if the query was successful
	return documents, nil
}

// GetDocument retrieves a document from the database based on its `idFile` field.
// It searches for a document with the given `idFile` and returns the document if found,
// or an error if not.
//
// Parameters:
// - idFile (uuid.UUID): The unique identifier of the document to retrieve.
//
// Returns:
// - *models.Document: A pointer to the document if found.
// - error: An error is returned if the document is not found or there is a database issue.
func GetDocument(idFile uuid.UUID) (*models.Document, error) {
	var document models.Document

	// Perform the query to find the document by its unique `idFile` field
	if err := config.DB.Where("deleted_at IS NULL AND id_file = ?", idFile).First(&document).Error; err != nil {
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

// GetDocumentByFingerprint retrieves a document from the database based on its unique fingerprint.
// It returns the document if found, or an error if not found or if any database-related issues occur.
//
// Parameters:
// - fingerprint (string): The unique fingerprint of the document to retrieve.
//
// Returns:
// - *models.Document: A pointer to the `Document` struct if the document is found.
// - error: An error if the document is not found or if there is a failure during the query.
func GetDocumentByFingerprint(fingerprint string) (*models.Document, error) {
	var document models.Document

	// Perform the query to find the document by its unique fingerprint
	if err := config.DB.Where("fingerprint = ?", fingerprint).First(&document).Error; err != nil {
		// If no record is found, return a descriptive error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("document with fingerprint %v not found", fingerprint)
		}
		// If there is another error during retrieval, return the error
		return nil, fmt.Errorf("error while retrieving document: %v", err)
	}

	// Return the document if found
	return &document, nil
}

// AddDocument adds a new document to the database.
// The function receives a pointer to a `Document` struct and attempts to insert it into the database.
//
// Parameters:
// - document (*models.Document): A pointer to the document to add to the database.
//
// Returns:
// - error: Returns an error if there is an issue during the insertion, or nil if successful.
func AddDocument(document *models.Document) error {
	// Create a new record for the document in the database
	if err := config.DB.Create(document).Error; err != nil {
		// If an error occurs during the insert, return the error
		return fmt.Errorf("error while adding document: %v", err)
	}
	// If the operation is successful, return nil (no error)
	return nil
}

// DeleteDocument deletes a document from the database by its associated idFile.
// This function searches for a document by `idFile`, and if found, deletes it from the database.
//
// Parameters:
// - idFile (uuid.UUID): The unique identifier of the document to delete.
//
// Returns:
// - error: Returns an error if the document is not found or if there is a failure during deletion.
func DeleteDocument(idFile uuid.UUID) error {
	// Declare a variable to hold the document from the database.
	var document models.Document

	// Retrieve the document using the provided idFile.
	// The 'Where' clause filters by the 'id_file' field.
	// 'First' retrieves the first matching record (if any).
	if err := config.DB.Where("id_file = ?", idFile).First(&document).Error; err != nil {
		// If the record is not found, return a custom error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("document with idFile %v not found", idFile)
		}
		// For any other error (e.g., database connection issues), return a generic error.
		return fmt.Errorf("error while fetching document: %v", err)
	}

	// If document is found, proceed to delete it.
	if err := config.DB.Delete(&document).Error; err != nil {
		// Return an error if the deletion failed.
		return fmt.Errorf("error while deleting document: %v", err)
	}

	// If no error occurred, return nil (indicating success).
	return nil
}
