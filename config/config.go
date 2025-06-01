package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Application struct {
	Server Server `json:"server"`
}

type Server struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

var App Application

const configDir = "config"

func Initialize(profile string) error {
	filename, err := checkProfileAndGetFilePath(profile)
	if err != nil {
		return fmt.Errorf("errore checking profile: %v", err)
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	if err = json.Unmarshal(content, &App); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}
	return nil
}

func checkProfileAndGetFilePath(profile string) (string, error) {
	var filename string
	switch {
	case profile == "dev":
		filename = fmt.Sprintf("%s/application-dev.json", configDir)
	case profile == "test":
		filename = fmt.Sprintf("%s/application-test.json", configDir)
	case profile == "prod":
		filename = fmt.Sprintf("%s/application.json", configDir)
	default:
		return "", fmt.Errorf("profile %s is not valid value", profile)
	}
	return filename, nil
}
