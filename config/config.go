package config

import (
	"encoding/json"
	"fileserver/internal/utils"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

// Application represents the top-level structure of the application's configuration.
type Application struct {
	Server   *Server   `json:"server"`   // Server configuration
	Database *Database `json:"database"` // Database configuration
	Minio    *Minio    `json:"minio"`    // MinIO configuration
}

// Server holds the configuration related to the web server (e.g., host, port).
type Server struct {
	Host string `json:"host"` // Hostname or IP address for the server
	Port int    `json:"port"` // Port number on which the server will run
}

// Database holds the configuration for connecting to a database (e.g., Postgres or SQLite).
type Database struct {
	Url      string `json:"url"`      // Database URL (used in case of SQLite)
	Driver   string `json:"driver"`   // Database driver (e.g., "postgres" or "sqlite")
	Host     string `json:"host"`     // Hostname of the database server (used in case of PostgreSQL)
	Port     int    `json:"port"`     // Port number for the database connection
	Name     string `json:"name"`     // Database name
	Username string `json:"username"` // Database username
	Password string `json:"password"` // Database password
	SSLMode  bool   `json:"ssl-mode"` // Whether SSL is enabled for the connection
	Timezone string `json:"timezone"` // Timezone for the database connection
}

// Minio holds the configuration for connecting to a MinIO server.
type Minio struct {
	Url          string `json:"url"`          // MinIO server URL
	Username     string `json:"username"`     // MinIO username
	Password     string `json:"password"`     // MinIO password
	Token        string `json:"token"`        // Optional token for MinIO authentication
	Secure       bool   `json:"secure"`       // Whether the connection is secure (HTTPS)
	Region       string `json:"region"`       // MinIO server region
	BucketLookup int    `json:"bucketLookup"` // Bucket lookup strategy
}

// Global variables for the application configuration and clients.
var (
	App   Application   // Application-level configuration
	DB    *gorm.DB      // Database client (GORM)
	MinIO *minio.Client // MinIO client
)

const (
	configDir = "config" // Directory where the configuration files are stored
)

// Initialize reads the configuration file based on the profile (dev, test, prod),
// and initializes the MinIO and database clients based on the configuration.
func Initialize(profile string) error {
	// Get the file path based on the profile
	filename, err := checkProfileAndGetFilePath(profile)
	if err != nil {
		return fmt.Errorf("error checking profile: %v", err)
	}

	// Read the configuration file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Unmarshal the JSON content into the Application structure
	if err = json.Unmarshal(content, &App); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	// Initialize MinIO if MinIO configuration is provided
	if App.Minio != nil {
		if err := initializeMinIO(App.Minio); err != nil {
			return fmt.Errorf("error initializing MinIO: %v", err)
		}
		fmt.Println("MinIO initialized")
	}

	// Initialize database if database configuration is provided
	if App.Database != nil {
		if err := initializeDatabase(App.Database); err != nil {
			return fmt.Errorf("error initializing database: %v", err)
		}
		fmt.Println("Database initialized")
	}

	return nil
}

// checkProfileAndGetFilePath returns the correct configuration file path based on the profile (dev, test, prod).
func checkProfileAndGetFilePath(profile string) (string, error) {
	var filename string
	// Match the profile to its respective configuration file
	switch {
	case profile == "dev":
		filename = fmt.Sprintf("%s/application-dev.json", configDir)
	case profile == "test":
		filename = fmt.Sprintf("%s/application-test.json", configDir)
	case profile == "prod":
		filename = fmt.Sprintf("%s/application.json", configDir)
	default:
		return "", fmt.Errorf("profile %s is not valid", profile)
	}
	return filename, nil
}

// initializeMinIO initializes the MinIO client using the provided configuration.
func initializeMinIO(minioConfig *Minio) error {
	// Create a MinIO client with the given credentials and options
	client, err := minio.New(minioConfig.Url, &minio.Options{
		Creds:        credentials.NewStaticV4(minioConfig.Username, minioConfig.Password, minioConfig.Token),
		Secure:       minioConfig.Secure,
		Region:       minioConfig.Region,
		BucketLookup: getBucketLookup(minioConfig.BucketLookup),
	})
	if err != nil {
		return fmt.Errorf("cannot connect to MinIO %s: %v", minioConfig.Url, err)
	}
	MinIO = client
	return nil
}

// getBucketLookup maps the integer value to the appropriate MinIO bucket lookup type.
func getBucketLookup(value int) minio.BucketLookupType {
	switch value {
	case 0:
		return minio.BucketLookupAuto
	case 1:
		return minio.BucketLookupDNS
	case 2:
		return minio.BucketLookupPath
	default:
		return minio.BucketLookupAuto
	}
}

// initializeDatabase initializes the database client based on the provided configuration.
func initializeDatabase(dbConfig *Database) error {
	// Generate the database connection string based on the driver
	switch dbConfig.Driver {
	case "postgres":
		var url string
		if dbConfig.Url != "" {
			url = fmt.Sprintf(
				"postgres://%s:%s@%s/%s?sslmode=%s&TimeZone=%s",
				dbConfig.Username,
				dbConfig.Password,
				dbConfig.Url,
				utils.DefaultValue(dbConfig.Name, "postgres"),
				getSSLModeValue(dbConfig.SSLMode),
				utils.DefaultValue(dbConfig.Timezone, "UTC"),
			)
		} else {
			url = fmt.Sprintf(
				"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
				dbConfig.Host,
				dbConfig.Username,
				dbConfig.Password,
				utils.DefaultValue(dbConfig.Name, "postgres"),
				dbConfig.Port,
				getSSLModeValue(dbConfig.SSLMode),
				utils.DefaultValue(dbConfig.Timezone, "UTC"),
			)
		}

		// Open PostgreSQL connection with GORM
		db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("cannot connect to database %s@%s:%d", dbConfig.Username, dbConfig.Host, dbConfig.Port)
		}
		DB = db
	case "sqlite":
		// Open SQLite connection with GORM
		db, err := gorm.Open(sqlite.Open(dbConfig.Url), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("cannot connect to database %s", dbConfig.Url)
		}
		DB = db
	default:
		return fmt.Errorf("database type is not supported")
	}
	return nil
}

// getSSLModeValue returns "enable" or "disable" based on the boolean value for SSL mode.
func getSSLModeValue(mode bool) string {
	if !mode {
		return "disable"
	}
	return "enable"
}
