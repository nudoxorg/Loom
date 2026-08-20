package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nudoxorg/loom/internal/project"
)

type Settings struct {
	DefaultAgent string `json:"default_agent"`
	DefaultLimit int    `json:"default_limit"`
}

// Default returns the settings Loom uses when no settings.json exists yet.
func Default() Settings {
	return Settings{
		DefaultAgent: "manual",
		DefaultLimit: 20,
	}
}

// Path returns ~/.loom/settings.json.
func Path() (string, error) {
	homeDir, err := project.HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, "settings.json"), nil
}

// Load reads settings.json, returning Default() if it doesn't exist yet.
func Load() (Settings, error) {
	path, err := Path()
	if err != nil {
		return Settings{}, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("reading settings: %w", err)
	}

	settings := Default()
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, fmt.Errorf("parsing settings: %w", err)
	}

	return settings, nil
}

// Save writes settings to settings.json.
func Save(s Settings) error {
	path, err := Path()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	return nil
}

// Keys lists the known settings keys, for `loom config list`.
func Keys() []string {
	return []string{"default_agent", "default_limit"}
}

// Get reads a single setting by key.
func Get(s Settings, key string) (string, error) {
	switch key {
	case "default_agent":
		return s.DefaultAgent, nil
	case "default_limit":
		return strconv.Itoa(s.DefaultLimit), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// Set updates a single setting by key, parsing value as needed.
func Set(s *Settings, key, value string) error {
	switch key {
	case "default_agent":
		s.DefaultAgent = value
	case "default_limit":
		limit, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("default_limit must be an integer: %w", err)
		}
		s.DefaultLimit = limit
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return nil
}
