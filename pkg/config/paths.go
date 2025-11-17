package config

import (
	"os"
	"path/filepath"
)

// GetConfigDir returns the config directory path (XDG_CONFIG_HOME)
func GetConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "lissto"), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

// GetCacheDir returns the cache directory path (XDG_CACHE_HOME)
func GetCacheDir() (string, error) {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheHome = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheHome, "lissto"), nil
}

// GetEnvCachePath returns the full path to the env cache file
func GetEnvCachePath() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "envs.yaml"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(configDir, 0755)
}

// EnsureCacheDir creates the cache directory if it doesn't exist
func EnsureCacheDir() error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(cacheDir, 0755)
}
