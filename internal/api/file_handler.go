package api

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

func LoadFile(w http.ResponseWriter, r *http.Request) {
	// Make sure the request is a POST and is of type multipart/form-data
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit the maximum request size (e.g., 10 MB)
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing the request", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the 'file' field of the form
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

	// Check the file extension (example)
	if !strings.HasSuffix(header.Filename, ".txt") {
		http.Error(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	// Create uploads folder if it doesn't exist
	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating the uploads folder", http.StatusInternalServerError)
		return
	}

	// Use a unique name for the file (timestamp)
	newFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	out, err := os.Create("./uploads/" + newFileName)
	if err != nil {
		http.Error(w, "Error saving the file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			http.Error(w, "Error closing the output file: "+err.Error(), http.StatusInternalServerError)
		}
	}(out)

	// Copy the file contents from the request to the saved file
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	fmt.Fprintf(w, "File %s uploaded successfully!", newFileName)
}
