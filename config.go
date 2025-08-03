package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type RepoConfig struct {
	DriveFolderID string    `json:"drive_folder_id"`
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token"`
	TokenExpiry   time.Time `json:"token_expiry"`
}

type Config struct {
	UserEmail string                `json:"user_email"`
	Repos     map[string]RepoConfig `json:"repos"`
}

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".gitdock", "config.json")
}

func LoadConfig() (Config, error) {
	path := getConfigPath()
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

func SaveConfig(config Config) error {
	path := getConfigPath()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
