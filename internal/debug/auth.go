package debug

import (
	"fmt"
	"net/url"
	"strings"

	"PostedIn/internal/config"
	"PostedIn/pkg/linkedin"
)

func ValidateLinkedInConfig(cfg *config.Config) error {
	if cfg.LinkedIn.ClientID == "" {
		return fmt.Errorf("LinkedIn Client ID is empty")
	}

	if cfg.LinkedIn.ClientSecret == "" {
		return fmt.Errorf("LinkedIn Client Secret is empty")
	}

	if cfg.LinkedIn.RedirectURL == "" {
		return fmt.Errorf("LinkedIn Redirect URL is empty")
	}

	// Validate redirect URL format
	parsedURL, err := url.Parse(cfg.LinkedIn.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid redirect URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("redirect URL must use http or https scheme")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("redirect URL must have a valid host")
	}

	return nil
}

func PrintAuthDetails(cfg *config.Config) {
	fmt.Println("🔍 LinkedIn Authentication Debug Info")
	fmt.Println("=====================================")
	fmt.Printf("Client ID: %s\n", maskString(cfg.LinkedIn.ClientID))
	fmt.Printf("Redirect URL: %s\n", cfg.LinkedIn.RedirectURL)

	// Create LinkedIn client and get auth URL
	linkedinConfig := linkedin.NewConfig(
		cfg.LinkedIn.ClientID,
		cfg.LinkedIn.ClientSecret,
		cfg.LinkedIn.RedirectURL,
	)
	client := linkedin.NewClient(linkedinConfig)
	authURL := client.GetAuthURL("linkedin-auth-state")

	fmt.Printf("Full Auth URL: %s\n", authURL)

	// Parse and validate the auth URL
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		fmt.Printf("❌ Error parsing auth URL: %v\n", err)
		return
	}

	fmt.Println("\n📋 Auth URL Components:")
	fmt.Printf("  Base URL: %s://%s%s\n", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	queryParams := parsedURL.Query()
	for key, values := range queryParams {
		if key == "client_id" {
			fmt.Printf("  %s: %s\n", key, maskString(values[0]))
		} else {
			fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
		}
	}

	// Validate required parameters
	fmt.Println("\n✅ Required Parameters Check:")
	checkParam(queryParams, "client_id", "Client ID")
	checkParam(queryParams, "redirect_uri", "Redirect URI")
	checkParam(queryParams, "response_type", "Response Type")
	checkParam(queryParams, "scope", "Scopes")
	checkParam(queryParams, "state", "State")
}

func checkParam(params url.Values, key, name string) {
	if values, exists := params[key]; exists && len(values) > 0 && values[0] != "" {
		if key == "client_id" {
			fmt.Printf("  ✓ %s: %s\n", name, maskString(values[0]))
		} else {
			fmt.Printf("  ✓ %s: %s\n", name, values[0])
		}
	} else {
		fmt.Printf("  ❌ %s: Missing or empty\n", name)
	}
}

func maskString(s string) string {
	const (
		maxLength = 8
		prefixLen = 4
		suffixLen = 4
		maskStr   = "****"
	)
	if len(s) <= maxLength {
		return maskStr
	}
	return s[:prefixLen] + maskStr + s[len(s)-suffixLen:]
}

func PrintCommonIssues() {
	fmt.Println("\n🚨 Common LinkedIn OAuth Issues:")
	fmt.Println("================================")
	fmt.Println("1. 'Network Will Be Back Soon' Error:")
	fmt.Println("   - Usually indicates invalid Client ID")
	fmt.Println("   - Check if Client ID is correct in config.json")
	fmt.Println("   - Verify the app exists in LinkedIn Developer Portal")
	fmt.Println()
	fmt.Println("2. 'Invalid redirect_uri' Error:")
	fmt.Println("   - Redirect URL in config.json must match LinkedIn app settings exactly")
	fmt.Println("   - Common format: http://localhost:8080/callback")
	fmt.Println()
	fmt.Println("3. 'App not found' Error:")
	fmt.Println("   - LinkedIn app may be deleted or suspended")
	fmt.Println("   - Check app status in LinkedIn Developer Portal")
	fmt.Println()
	fmt.Println("4. Scope Issues:")
	fmt.Println("   - Ensure 'openid', 'profile', and 'w_member_social' scopes are enabled")
	fmt.Println("   - Some scopes may require LinkedIn review")
	fmt.Println()
	fmt.Println("💡 Next Steps:")
	fmt.Println("1. Verify your LinkedIn app exists and is active")
	fmt.Println("2. Double-check Client ID and Client Secret")
	fmt.Println("3. Ensure redirect URL matches exactly")
	fmt.Println("4. Try creating a new LinkedIn app if issues persist")
}
