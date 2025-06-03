package utils

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
)

// CalculateFingerprint calculates the fingerprint (SHA-1 hash) of a file at a given path.
//
// This function computes a SHA-1 hash for the contents of a file. It opens the file, reads it in chunks
// to avoid loading the entire file into memory, and then calculates the hash using the SHA-1 algorithm.
// Finally, it returns the resulting hash as a hexadecimal string.
//
// Parameters:
//   - filePath (string): The path to the file whose fingerprint (hash) is to be calculated.
//
// Returns:
//   - string: The SHA-1 hash of the file, represented as a hexadecimal string.
//   - error: Any error encountered while opening the file, reading it, or calculating the hash.
//     If no error occurred, it returns nil.
//
// Example usage:
//
//	fingerprint, err := utils.CalculateFingerprint("/path/to/file.txt")
//	if err != nil {
//	    log.Fatalf("Error calculating fingerprint: %v", err)
//	}
//	fmt.Printf("Fingerprint: %s\n", fingerprint)
func CalculateFingerprint(filePath string) (string, error) {
	// Open the file at the specified file path.
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	// Ensure that the file is closed after processing (using a defer statement).
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close file: %v", err)
		}
	}(file)

	// Create a new SHA-1 hash object.
	hash := sha1.New()

	// Read the file and calculate the hash while reading. The entire file is not loaded into memory.
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", fmt.Errorf("failed to calculate hash: %v", err)
	}

	// Return the final hash as a hexadecimal string.
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
