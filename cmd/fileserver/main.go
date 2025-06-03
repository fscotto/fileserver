package main

import (
	"fileserver/config"
	"fileserver/internal/api"
	"fileserver/internal/utils"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Defer a function to catch any runtime panics and log them.
	// This helps in recovering from unexpected fatal errors.
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Catch fatal error: %v\n", r) // Log the fatal error if panic occurs
		}
	}()

	// Get the application profile from environment variables or default to "prod" if not set.
	profile := utils.DefaultValue(os.Getenv("APP_PROFILE"), "prod")

	// Initialize the configuration for the application based on the profile.
	if err := config.Initialize(profile); err != nil {
		// If an error occurs during initialization, log the error and terminate the application.
		log.Fatalf("Error to read %s configuration: %v\n", profile, err)
	}

	// Log the profile that is being used to start the application.
	log.Printf("Application starting with profile: %s", profile)

	// Create a new HTTP request multiplexer (ServeMux) to register routes.
	mux := http.NewServeMux()
	log.Printf("Register all routes\n")

	// Iterate through the routes defined in the API package and register them.
	for url, handler := range api.Routes {
		// For each route, log the URL and corresponding handler function name.
		log.Printf("Register route %s for %v", url, utils.GetFunctionName(handler))
		// Register the route and associate it with the handler function.
		mux.HandleFunc(url, handler)
	}

	// Get the server configuration from the app's config settings.
	server := config.App.Server
	// Format the server's host and port into a string for the URL.
	url := fmt.Sprintf("%s:%d", server.Host, server.Port)
	// Log the server's URL where it will be listening.
	log.Printf("Start server on %s\n", url)

	// Start the HTTP server using the specified host and port, and pass in the mux for routing.
	// If an error occurs while starting the server, log it and terminate the program.
	if err := http.ListenAndServe(url, mux); err != nil {
		log.Fatalf("%v\n", err)
	}
}
