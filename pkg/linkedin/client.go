// Package linkedin provides LinkedIn API client functionality for OAuth authentication and post publishing.
package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("[linkedin] No .env file found or failed to load .env (this is OK if running in production):", err)
	}
}

const (
	httpTimeout = 30 * time.Second
)

const (
	// AuthURL is the LinkedIn OAuth authorization endpoint.
	AuthURL = "oauth/v2/authorization"
	// TokenURL is the LinkedIn OAuth token exchange endpoint.
	TokenURL = "oauth/v2/accessToken"
	// UserInfoURL is the LinkedIn user info endpoint.
	UserInfoURL = "v2/userinfo"
	// APIBaseURL is the base URL for LinkedIn API v2.
	APIBaseURL = "rest"
	// PostsURL is the LinkedIn posts API endpoint.
	PostsURL = APIBaseURL + "/posts"
	// ImagesURL is the LinkedIn images API endpoint.
	ImagesURL = APIBaseURL + "/images"
)

// Config holds LinkedIn OAuth configuration parameters.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// Client provides LinkedIn API functionality with OAuth authentication.
type Client struct {
	config *oauth2.Config
	token  *oauth2.Token
	client *http.Client
}

// Post represents a LinkedIn post structure for API requests.
type Post struct {
	Author                    string                 `json:"author"`
	Commentary                string                 `json:"commentary"`
	Visibility                string                 `json:"visibility"`
	Distribution              map[string]interface{} `json:"distribution"`
	LifecycleState            string                 `json:"lifecycleState"`
	IsReshareDisabledByAuthor bool                   `json:"isReshareDisabledByAuthor,omitempty"`
	Content                   *Content               `json:"content,omitempty"`
}

// Content represents the content field for a LinkedIn post, supporting both single and multi-image.
type Content struct {
	Media      *Media      `json:"media,omitempty"`
	MultiImage *MultiImage `json:"multiImage,omitempty"`
}

// Media represents a single image for a LinkedIn post.
type Media struct {
	ID      string `json:"id"`
	AltText string `json:"altText"`
}

// MultiImage represents the multiImage field for a LinkedIn post.
type MultiImage struct {
	Images []Image `json:"images"`
}

// Image represents an image in a multi-image LinkedIn post.
type Image struct {
	ID      string `json:"id"`
	AltText string `json:"altText"`
}

// NewConfig creates a new LinkedIn OAuth configuration.
func NewConfig(clientID, clientSecret, redirectURL string) *Config {
	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "w_member_social", "email"},
	}
}

// NewClient creates a new LinkedIn API client with the given configuration.
func NewClient(config *Config) *Client {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/%s", os.Getenv("LINKEDIN_BASE_URL"), AuthURL),
			TokenURL: fmt.Sprintf("%s/%s", os.Getenv("LINKEDIN_BASE_URL"), TokenURL),
		},
	}

	return &Client{
		config: oauth2Config,
		client: &http.Client{},
	}
}

// GetAuthURL generates the OAuth authorization URL for LinkedIn.
func (c *Client) GetAuthURL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeToken exchanges an authorization code for an access token.
func (c *Client) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	c.token = token
	c.client = c.config.Client(ctx, token)

	return token, nil
}

// SetToken sets the OAuth access token for the client.
func (c *Client) SetToken(token *oauth2.Token) {
	c.token = token
	c.client = c.config.Client(context.Background(), token)
}

// GetProfile retrieves the LinkedIn user profile information.
func (c *Client) GetProfile(ctx context.Context) (map[string]interface{}, error) {
	if c.token == nil {
		return nil, fmt.Errorf("no access token available")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s", os.Getenv("LINKEDIN_API_URL"), UserInfoURL), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PostedIn/1.0")
	req.Header.Set("LinkedIn-Version", "202506")

	client := &http.Client{
		Timeout: httpTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Warning: failed to close response body: %v\n", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var profile map[string]interface{}
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return profile, nil
}

// CreatePost creates a new LinkedIn post with the given text content and optional images.
func (c *Client) CreatePost(ctx context.Context, text, userID string, images []Image) error {
	if c.token == nil {
		return fmt.Errorf("no access token available")
	}

	post := Post{
		Author:     "urn:li:person:" + userID,
		Commentary: text,
		Visibility: "PUBLIC",
		Distribution: map[string]interface{}{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []interface{}{},
			"thirdPartyDistributionChannels": []interface{}{},
		},
		LifecycleState: "PUBLISHED",
	}

	if len(images) == 1 {
		post.Content = &Content{
			Media: &Media{
				ID:      images[0].ID,
				AltText: images[0].AltText,
			},
		}
		post.IsReshareDisabledByAuthor = false
	} else if len(images) >= 2 {
		post.Content = &Content{
			MultiImage: &MultiImage{
				Images: images,
			},
		}
		post.IsReshareDisabledByAuthor = false
	}

	// Debug: print the post payload
	fmt.Printf("DEBUG: Creating post with author: %s\n", post.Author)
	fmt.Printf("DEBUG: User ID: %s\n", userID)
	if post.Content != nil {
		fmt.Printf("DEBUG: Images: %+v\n", images)
	}

	jsonData, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s", os.Getenv("LINKEDIN_API_URL"), PostsURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PostedIn/1.0")
	req.Header.Set("LinkedIn-Version", "202506")

	client := &http.Client{
		Timeout: httpTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Warning: failed to close response body: %v\n", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UploadImage uploads an image to LinkedIn using the new Images API and returns the image URN.
func (c *Client) UploadImage(ctx context.Context, imageData []byte, fileName, userURN string) (string, error) {
	if c.token == nil {
		return "", fmt.Errorf("no access token available")
	}

	// 1. Initialize the upload (new Images API)
	type initializeUploadRequest struct {
		InitializeUploadRequest struct {
			Owner string `json:"owner"`
		} `json:"initializeUploadRequest"`
	}

	initReq := initializeUploadRequest{}
	initReq.InitializeUploadRequest.Owner = userURN

	jsonData, err := json.Marshal(initReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal initialize upload request: %w", err)
	}

	uploadURL := fmt.Sprintf("%s/%s?action=initializeUpload", os.Getenv("LINKEDIN_API_URL"), ImagesURL)

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create initialize upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PostedIn/1.0")
	req.Header.Set("LinkedIn-Version", "202506")
	req.Header.Set("X-RestLi-Protocol-Version", "2.0.0")

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to initialize upload: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Warning: failed to close response body: %v\n", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read initialize upload response: %w", err)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("initialize upload API error (%d): %s", resp.StatusCode, string(body))
	}

	var initResp struct {
		Value struct {
			UploadUrl string `json:"uploadUrl"`
			Image     string `json:"image"`
		} `json:"value"`
	}
	if err := json.Unmarshal(body, &initResp); err != nil {
		return "", fmt.Errorf("failed to parse initialize upload response: %w", err)
	}

	uploadTo := initResp.Value.UploadUrl
	imageURN := initResp.Value.Image
	if uploadTo == "" || imageURN == "" {
		return "", fmt.Errorf("missing upload URL or image URN in response")
	}

	// 2. Upload the image bytes (PUT)
	putReq, err := http.NewRequestWithContext(ctx, "PUT", uploadTo, bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to create image upload request: %w", err)
	}
	// Set content type based on file extension (default to image/jpeg)
	contentType := "image/jpeg"
	if fileName != "" {
		if ext := getFileExtension(fileName); ext != "" {
			switch ext {
			case ".png":
				contentType = "image/png"
			case ".gif":
				contentType = "image/gif"
			case ".bmp":
				contentType = "image/bmp"
			case ".webp":
				contentType = "image/webp"
			}
		}
	}
	putReq.Header.Set("Content-Type", contentType)

	putResp, err := client.Do(putReq)
	if err != nil {
		return "", fmt.Errorf("failed to upload image bytes: %w", err)
	}
	defer func() {
		if cerr := putResp.Body.Close(); cerr != nil {
			log.Printf("Warning: failed to close put response body: %v\n", cerr)
		}
	}()
	if putResp.StatusCode != http.StatusCreated && putResp.StatusCode != http.StatusOK {
		putBody, _ := io.ReadAll(putResp.Body)
		return "", fmt.Errorf("image upload error (%d): %s", putResp.StatusCode, string(putBody))
	}

	// 3. Return the image URN
	return imageURN, nil
}

// GetImageDownloadURL fetches the downloadUrl for a LinkedIn image URN.
func (c *Client) GetImageDownloadURL(ctx context.Context, urn string) (string, error) {
	if c.token == nil {
		return "", fmt.Errorf("no access token available")
	}
	urn = strings.TrimSpace(urn)
	endpoint := fmt.Sprintf("%s/%s/%s", os.Getenv("LINKEDIN_API_URL"), ImagesURL, urn)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("LinkedIn-Version", "202506")

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get image: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Warning: failed to close response body: %v\n", cerr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}
	var result struct {
		DownloadUrl string `json:"downloadUrl"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if result.DownloadUrl == "" {
		return "", fmt.Errorf("no downloadUrl in response")
	}
	return result.DownloadUrl, nil
}

// getFileExtension returns the lowercase file extension (including dot), or empty string.
func getFileExtension(fileName string) string {
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return fileName[i:]
		}
	}
	return ""
}

// IsAuthenticated checks if the client has a valid access token.
func (c *Client) IsAuthenticated() bool {
	return c.token != nil && c.token.Valid()
}
