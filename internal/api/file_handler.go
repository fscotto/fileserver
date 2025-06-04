package api

import (
	"context"
	"encoding/json"
	"fileserver/config"
	"fileserver/internal/models"
	"fileserver/internal/service"
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// Constants for default bucket name and folder path for uploaded files
const (
	defaultBucketName   = "documents"              // Default bucket name in MinIO
	localFolderTemplate = "%s/fileserver/uploads/" // Template for creating local upload directories
)

// GetFiles retrieves the list of indexed documents from the database with fuzzy search on file names
func GetFiles(w http.ResponseWriter, r *http.Request) {
	// Step 1: Retrieve the search query from the URL parameters
	searchQuery := r.URL.Query().Get("searchQuery")
	if searchQuery == "" {
		// If there is no search query, retrieve all documents
		searchQuery = "%"
	} else {
		// Add wildcards for partial search
		searchQuery = "%" + searchQuery + "%"
	}

	// Step 2: Retrieve documents whose name matches the fuzzy search
	documents, err := service.GetFiles(searchQuery)
	if err != nil {
		// Handle error if the query fails
		http.Error(w, fmt.Sprintf("Error retrieving documents: %v", err), http.StatusInternalServerError)
		return
	}

	// Step 3: Convert the documents to JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Use json.NewEncoder to write the response directly in JSON format
	if err := json.NewEncoder(w).Encode(documents); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetFile handles the request to fetch a file from MinIO and serve it to the user.
func GetFile(w http.ResponseWriter, r *http.Request) {
	// Ensure that the request method is GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the object name from the query parameters
	objectName := r.PathValue("idFile")
	idFile, err := uuid.Parse(objectName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing objectName: %v", err), http.StatusBadRequest)
		return
	}

	document, err := service.GetDocument(idFile)
	if document == nil || err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving document: %v", err), http.StatusNotFound)
		return
	}

	// Fetch the file object from MinIO storage
	object, err := service.GetFileFromMinIO(defaultBucketName, objectName)
	if err != nil {
		return
	}
	defer object.Close() // Ensure that the file object is closed after use

	// Create the upload directory if it doesn't exist
	uploadDir := fmt.Sprintf(localFolderTemplate, os.TempDir())
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating the uploads folder", http.StatusInternalServerError)
		return
	}

	// Create a new file locally with a unique name (using a timestamp)
	newFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), objectName)
	file, err := os.Create(uploadDir + newFileName)
	if err != nil {
		http.Error(w, "Error saving the file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			http.Error(w, "Error closing file: "+err.Error(), http.StatusInternalServerError)
		}
	}(file)

	// Copy the file content from MinIO to the local file
	_, err = io.Copy(file, object)
	if err != nil {
		fmt.Println("Error saving object to file:", err)
		return
	}

	// Get the file's information (size, name, etc.)
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not get file information", http.StatusInternalServerError)
		return
	}

	// Set headers for file download (name, content type, and length)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; objectName=%s", fileInfo.Name()))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Log the successful file retrieval
	fmt.Printf("Sending file: %s (Size: %d bytes)\n", fileInfo.Name(), fileInfo.Size())

	// Serve the file content as a download
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

// LoadFile handles file uploads from a client and stores them locally and on MinIO.
func LoadFile(w http.ResponseWriter, r *http.Request) {
	// Ensure that the request method is POST and that it is a multipart form
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form data with a maximum file size of 10MB
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing the request", http.StatusBadRequest)
		return
	}

	// Retrieve the uploaded file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			http.Error(w, "Error closing the input file: "+err.Error(), http.StatusInternalServerError)
		}
	}(file)

	// Create the upload directory if it doesn't exist
	uploadDir := fmt.Sprintf(localFolderTemplate, os.TempDir())
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating the uploads folder", http.StatusInternalServerError)
		return
	}

	// Use a unique file name based on timestamp and the original file name
	newFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filePath := uploadDir + newFileName
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving the file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			http.Error(w, "Error closing the output file: "+err.Error(), http.StatusInternalServerError)
		}
		if err := cleanup(out); err != nil {
			http.Error(w, "Error remove the output file: "+err.Error(), http.StatusInternalServerError)
		}
	}(out)

	// Copy the file content from the request to the local file
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
		return
	}

	// Calculate fingerprint of file
	//fingerprint, err := utils.CalculateFingerprint(filePath)
	//if err != nil {
	//	http.Error(w, "Error during calculate fingerprint: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	fingerprint := uuid.New().String()

	// Check if document already uploaded
	_, err = service.GetDocumentByFingerprint(fingerprint)
	if err == nil {
		http.Error(w, "Document already exists.", http.StatusConflict)
		return
	}

	// Upload the file to MinIO with a unique ID (UUID)
	idFile := uuid.New()
	err = service.UploadFileToMinIO(context.Background(), defaultBucketName, idFile.String(), filePath)
	if err != nil {
		http.Error(w, "Error during upload file to MinIO: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save document to database
	newDocument := &models.Document{
		Name:        header.Filename,
		IdFile:      idFile,
		Fingerprint: fingerprint,
	}
	if err := service.AddDocument(newDocument); err != nil {
		http.Error(w, "Error adding document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond to the client with a success message
	_, err = fmt.Fprintf(w, "File %s uploaded successfully!\n", newFileName)
	if err != nil {
		return
	}
}

// cleanup removes a file from the local file system after use.
func cleanup(file *os.File) error {
	if _, err := os.Stat(file.Name()); err == nil {
		err := os.Remove(file.Name())
		if err != nil {
			return fmt.Errorf("Error removing file: %v", err)
		} else {
			fmt.Println("File removed successfully:", file.Name())
		}
	} else if os.IsNotExist(err) {
		return fmt.Errorf("File does not exist: %v", file.Name())
	} else {
		fmt.Println("Error checking file:", err)
	}
	return nil
}

// DeleteFile deletes a file from the database and MinIO
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	// Ensure that the request method is GET
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the object name from the path value
	idFile, err := uuid.Parse(r.PathValue("idFile"))
	if err != nil {
		http.Error(w, "Error parsing the idFile: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Step 1: Get document from PostgreSQL database
	document, err := service.GetDocument(idFile)
	if err != nil {
		http.Error(w, "Document not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Step 2: Delete from PostgreSQL
	if err := config.DB.Delete(&document).Error; err != nil {
		http.Error(w, fmt.Sprintf("Error deleting document from DB: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File with ID %v deleted successfully", idFile)
}
