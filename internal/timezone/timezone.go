// Package timezone provides timezone management and conversion utilities for the LinkedIn Post Scheduler.
package timezone

import (
	"fmt"
	"time"
)

const (
	secondsPerHour   = 3600
	secondsPerMinute = 60
)

// DetectLocalTimezone detects the system's local timezone.
func DetectLocalTimezone() (location, offset string, err error) {
	now := time.Now()
	zone, offsetSeconds := now.Zone()

	// Get the timezone location name
	location = now.Location().String()

	// Format offset as +/-HHMM to +/-HH:MM
	hours := offsetSeconds / secondsPerHour
	minutes := (offsetSeconds % secondsPerHour) / secondsPerMinute

	var offsetStr string
	if offsetSeconds >= 0 {
		offsetStr = fmt.Sprintf("+%02d:%02d", hours, minutes)
	} else {
		offsetStr = fmt.Sprintf("-%02d:%02d", -hours, -minutes)
	}

	// If location is "Local", try to get a better name
	if location == "Local" {
		location = getLocationFromZone(zone, offsetSeconds)
	}

	return location, offsetStr, nil
}

// getLocationFromZone attempts to map timezone abbreviations to locations.
func getLocationFromZone(zone string, offset int) string {
	// Common timezone mappings
	timezoneMap := map[string]map[int]string{
		"WIB":  {7 * secondsPerHour: "Asia/Jakarta"},         // Western Indonesian Time
		"WITA": {8 * secondsPerHour: "Asia/Makassar"},        // Central Indonesian Time
		"WIT":  {9 * secondsPerHour: "Asia/Jayapura"},        // Eastern Indonesian Time
		"ICT":  {7 * secondsPerHour: "Asia/Bangkok"},         // Indochina Time
		"JST":  {9 * secondsPerHour: "Asia/Tokyo"},           // Japan Standard Time
		"KST":  {9 * secondsPerHour: "Asia/Seoul"},           // Korea Standard Time
		"CST":  {8 * secondsPerHour: "Asia/Shanghai"},        // China Standard Time
		"SGT":  {8 * secondsPerHour: "Asia/Singapore"},       // Singapore Time
		"MYT":  {8 * secondsPerHour: "Asia/Kuala_Lumpur"},    // Malaysia Time
		"PHT":  {8 * secondsPerHour: "Asia/Manila"},          // Philippines Time
		"EST":  {-5 * secondsPerHour: "America/New_York"},    // Eastern Standard Time
		"PST":  {-8 * secondsPerHour: "America/Los_Angeles"}, // Pacific Standard Time
		"GMT":  {0: "Europe/London"},                         // Greenwich Mean Time
		"UTC":  {0: "UTC"},                                   // Coordinated Universal Time
	}

	if locations, exists := timezoneMap[zone]; exists {
		if location, exists := locations[offset]; exists {
			return location
		}
	}

	// Fallback: construct a generic location based on offset
	hours := offset / secondsPerHour
	if hours >= 0 {
		return fmt.Sprintf("Etc/GMT-%d", hours)
	}

	return fmt.Sprintf("Etc/GMT+%d", -hours)
}

// GetCommonTimezones returns a list of commonly used timezones.
func GetCommonTimezones() []Info {
	return []Info{
		{Name: "Asia/Jakarta", Description: "Western Indonesian Time (WIB)", Offset: "+07:00"},
		{Name: "Asia/Makassar", Description: "Central Indonesian Time (WITA)", Offset: "+08:00"},
		{Name: "Asia/Jayapura", Description: "Eastern Indonesian Time (WIT)", Offset: "+09:00"},
		{Name: "Asia/Bangkok", Description: "Thailand Time (ICT)", Offset: "+07:00"},
		{Name: "Asia/Singapore", Description: "Singapore Time (SGT)", Offset: "+08:00"},
		{Name: "Asia/Kuala_Lumpur", Description: "Malaysia Time (MYT)", Offset: "+08:00"},
		{Name: "Asia/Manila", Description: "Philippines Time (PHT)", Offset: "+08:00"},
		{Name: "Asia/Tokyo", Description: "Japan Standard Time (JST)", Offset: "+09:00"},
		{Name: "Asia/Seoul", Description: "Korea Standard Time (KST)", Offset: "+09:00"},
		{Name: "Asia/Shanghai", Description: "China Standard Time (CST)", Offset: "+08:00"},
		{Name: "America/New_York", Description: "Eastern Time (EST/EDT)", Offset: "-05:00/-04:00"},
		{Name: "America/Los_Angeles", Description: "Pacific Time (PST/PDT)", Offset: "-08:00/-07:00"},
		{Name: "America/Chicago", Description: "Central Time (CST/CDT)", Offset: "-06:00/-05:00"},
		{Name: "Europe/London", Description: "Greenwich Mean Time (GMT/BST)", Offset: "+00:00/+01:00"},
		{Name: "Europe/Paris", Description: "Central European Time (CET/CEST)", Offset: "+01:00/+02:00"},
		{Name: "UTC", Description: "Coordinated Universal Time", Offset: "+00:00"},
	}
}

// Info represents timezone information.
type Info struct {
	Name        string
	Description string
	Offset      string
}

// ValidateTimezone checks if a timezone location string is valid.
func ValidateTimezone(location string) error {
	_, err := time.LoadLocation(location)
	if err != nil {
		return fmt.Errorf("invalid timezone location '%s': %w", location, err)
	}

	return nil
}

// FormatTimezoneInfo returns formatted timezone information for display.
func FormatTimezoneInfo(location string) (string, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		return "", err
	}

	now := time.Now().In(loc)
	zone, offset := now.Zone()

	hours := offset / secondsPerHour
	minutes := (offset % secondsPerHour) / secondsPerMinute

	var offsetStr string
	if offset >= 0 {
		offsetStr = fmt.Sprintf("+%02d:%02d", hours, minutes)
	} else {
		offsetStr = fmt.Sprintf("-%02d:%02d", -hours, -minutes)
	}

	return fmt.Sprintf("%s (%s %s)", location, zone, offsetStr), nil
}

// GetCurrentTimeInTimezone returns the current time in the specified timezone.
func GetCurrentTimeInTimezone(location string) (time.Time, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		return time.Time{}, err
	}

	return time.Now().In(loc), nil
}
