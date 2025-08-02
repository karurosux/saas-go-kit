package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ProjectConfig struct {
	Version    string            `json:"version"`
	CliVersion string            `json:"cli_version"`
	CreatedAt  time.Time         `json:"created_at"`
	Project    struct {
		Name     string `json:"name"`
		GoModule string `json:"go_module"`
		Database string `json:"database"`
	} `json:"project"`
	Modules map[string]ModuleInfo `json:"modules"`
}

type ModuleInfo struct {
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installed_at"`
}

const configFileName = "sgk.json"

func InitProject() error {
	return InitProjectWithConfig("", "", "postgres")
}

func InitProjectWithConfig(projectName, goModule, database string) error {
	if _, err := os.Stat(configFileName); err == nil {
		return fmt.Errorf("project already initialized (sgk.json exists)")
	}

	if projectName == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectName = filepath.Base(pwd)
	}

	if goModule == "" {
		goModule = projectName
	}

	if database == "" {
		database = "postgres"
	}

	config := ProjectConfig{
		Version:    "1.0.0",
		CliVersion: "1.0.0",
		CreatedAt:  time.Now(),
		Modules:    make(map[string]ModuleInfo),
	}

	config.Project.Name = projectName
	config.Project.GoModule = goModule
	config.Project.Database = database

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFileName, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func LoadProjectConfig() (*ProjectConfig, error) {
	data, err := os.ReadFile(configFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w (run 'sgk init' first)", err)
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func SaveProjectConfig(config *ProjectConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFileName, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}