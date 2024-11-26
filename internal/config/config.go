package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// InstallerConfig represents the YAML configuration structure
type InstallerConfig struct {
	ToolList []string               `yaml:"tool_list"`
	Tools    map[string]*ToolConfig `yaml:"tools"`
}

// ToolConfig represents a tool's configuration
type ToolConfig struct {
	Dependencies []string        `yaml:"dependencies"`
	Version      string         `yaml:"version"`
	VersionFlag  string         `yaml:"version_flag"`
	Methods      []InstallMethod `yaml:"methods"`
}

// InstallMethod represents an installation method
type InstallMethod struct {
	Name     string   `yaml:"name"`
	Commands []string `yaml:"commands"`
}

// LoadConfig loads the installer configuration from a YAML file
func LoadConfig(filename string) (*InstallerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config InstallerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}
