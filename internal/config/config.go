// Package config provides application configuration management for LinkedIn Post Scheduler.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"PostedIn/internal/timezone"

	"golang.org/x/oauth2"
)

const (
	secondsPerHour   = 3600
	secondsPerMinute = 60
	restrictedPerm   = 0o600
)

// Config represents the main application configuration structure.
type Config struct {
	LinkedIn LinkedInConfig `json:"linkedin"`
	Storage  StorageConfig  `json:"storage"`
	Timezone TimezoneConfig `json:"timezone"`
	Cron     CronConfig     `json:"cron"`
}

// LinkedInConfig holds LinkedIn OAuth configuration settings.
type LinkedInConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
	UserID       string `json:"user_id,omitempty"`
}

// StorageConfig defines file paths for data storage.
type StorageConfig struct {
	PostsFile string `json:"posts_file"`
	TokenFile string `json:"token_file"`
}

// TimezoneConfig specifies timezone settings for post scheduling.
type TimezoneConfig struct {
	Location string `json:"location"`
	Offset   string `json:"offset"`
}

// CronConfig controls automatic post scheduling functionality.
type CronConfig struct {
	Enabled bool `json:"enabled"`
}

const (
	BaseConfigPath = "./internal/config"
	// ConfigFile is the default configuration file name.
	ConfigFile = BaseConfigPath + "/config.json"
	// TokenFile is the default OAuth token file name.
	TokenFile = BaseConfigPath + "/linkedin_token.json"
)

// LoadConfig loads application configuration from the config file or creates default configuration.
func LoadConfig() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		// Detect local timezone
		localLocation, localOffset, err := timezone.DetectLocalTimezone()
		if err != nil {
			// Fallback to Asia/Bangkok if detection fails
			localLocation = "Asia/Bangkok"
			localOffset = "+07:00"
		}

		// Create default config with local timezone
		defaultConfig := &Config{
			LinkedIn: LinkedInConfig{
				ClientID:     "",
				ClientSecret: "",
				RedirectURL:  "http://localhost:8080/callback",
			},
			Storage: StorageConfig{
				PostsFile: "posts.json",
				TokenFile: TokenFile,
			},
			Timezone: TimezoneConfig{
				Location: localLocation,
				Offset:   localOffset,
			},
			Cron: CronConfig{
				Enabled: true,
			},
		}

		if err := SaveConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}

		return nil, fmt.Errorf("config file created at %s with local timezone (%s %s) - please fill in your LinkedIn app credentials", ConfigFile, localLocation, localOffset)
	}

	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if config.LinkedIn.ClientID == "" || config.LinkedIn.ClientSecret == "" {
		return nil, fmt.Errorf("LinkedIn client_id and client_secret are required in %s", ConfigFile)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the config file.
func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(ConfigFile, data, restrictedPerm)
}

// LoadToken loads an OAuth token from the specified file.
func LoadToken(filename string) (*oauth2.Token, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("token file does not exist: %s", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}

// SaveToken saves an OAuth token to the specified file.
func SaveToken(token *oauth2.Token, filename string) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	return os.WriteFile(filename, data, restrictedPerm) // More restrictive permissions for token
}

// GetTimezone returns the configured timezone location.
func (c *Config) GetTimezone() (*time.Location, error) {
	if c.Timezone.Location == "" {
		// Default to UTC+7 if not configured
		return time.LoadLocation("Asia/Bangkok")
	}

	return time.LoadLocation(c.Timezone.Location)
}

// Now returns the current time in the configured timezone.
func (c *Config) Now() (time.Time, error) {
	loc, err := c.GetTimezone()
	if err != nil {
		return time.Time{}, err
	}

	return time.Now().In(loc), nil
}

// ParseTimeInTimezone parses a time string and returns it in the configured timezone.
func (c *Config) ParseTimeInTimezone(dateStr, timeStr string) (time.Time, error) {
	loc, err := c.GetTimezone()
	if err != nil {
		return time.Time{}, err
	}

	dateTimeStr := dateStr + " " + timeStr

	parsedTime, err := time.ParseInLocation("2006-01-02 15:04", dateTimeStr, loc)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

// SetDefaultTimezoneIfEmpty sets default timezone configuration if missing.
func (c *Config) SetDefaultTimezoneIfEmpty() {
	if c.Timezone.Location == "" {
		c.Timezone.Location = "Asia/Bangkok"
		c.Timezone.Offset = "+07:00"
	}
}

// UpdateTimezone updates the timezone configuration.
func (c *Config) UpdateTimezone(location string) error {
	// Validate the timezone
	if err := timezone.ValidateTimezone(location); err != nil {
		return err
	}

	// Get the current offset for the new timezone
	loc, err := time.LoadLocation(location)
	if err != nil {
		return fmt.Errorf("failed to load timezone: %w", err)
	}

	now := time.Now().In(loc)
	_, offset := now.Zone()

	hours := offset / secondsPerHour
	minutes := (offset % secondsPerHour) / secondsPerMinute

	var offsetStr string
	if offset >= 0 {
		offsetStr = fmt.Sprintf("+%02d:%02d", hours, minutes)
	} else {
		offsetStr = fmt.Sprintf("-%02d:%02d", -hours, -minutes)
	}

	// Update config
	c.Timezone.Location = location
	c.Timezone.Offset = offsetStr

	return nil
}

// GetTimezoneInfo returns formatted timezone information.
func (c *Config) GetTimezoneInfo() (string, error) {
	return timezone.FormatTimezoneInfo(c.Timezone.Location)
}

// GetCommonTimezones returns commonly used timezones.
func GetCommonTimezones() []timezone.Info {
	return timezone.GetCommonTimezones()
}

// DetectLocalTimezone returns the system's local timezone.
func DetectLocalTimezone() (location, offset string, err error) {
	return timezone.DetectLocalTimezone()
}
