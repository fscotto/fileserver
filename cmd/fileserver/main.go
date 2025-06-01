package main

import (
	"fileserver/config"
	"fileserver/internal/api"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Catch fatal error: %v\n", r)
		}
	}()
	profile := defaultValue(os.Getenv("APP_PROFILE"), "prod")
	if err := config.Initialize(profile); err != nil {
		log.Fatalf("Error to read %s configuration: %v\n", profile, err)
	}
	log.Printf("Application starting with profile: %s", profile)
	// Register all routes in the new ServeMux
	mux := http.NewServeMux()
	log.Printf("Register all routes\n")
	for url, handler := range api.Routes {
		log.Printf("Register route %s for %v", url, getFunctionName(handler))
		mux.HandleFunc(url, handler)
	}
	server := config.App.Server
	url := fmt.Sprintf("%s:%d", server.Host, server.Port)
	log.Printf("Start server on %s\n", url)
	if err := http.ListenAndServe(url, mux); err != nil {
		log.Fatalf("%v\n", err)
	}
}

// Function to get the name of a function from its reference
func getFunctionName(fn any) string {
	// Get the pointer to the function using reflection
	pc := reflect.ValueOf(fn).Pointer()

	// Use the pointer to get the function object
	funcObj := runtime.FuncForPC(pc)

	// Return the name of the function
	return funcObj.Name()
}

func defaultValue(value string, other string) string {
	if value != "" {
		return value
	}
	return other
}
