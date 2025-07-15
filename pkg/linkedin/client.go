package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	httpTimeout = 30 * time.Second
)

const (
	AuthURL     = "https://www.linkedin.com/oauth/v2/authorization"
	TokenURL    = "https://www.linkedin.com/oauth/v2/accessToken"
	UserInfoURL = "https://api.linkedin.com/v2/userinfo"
	APIBaseURL  = "https://api.linkedin.com/rest"
	PostsURL    = APIBaseURL + "/posts"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type Client struct {
	config *oauth2.Config
	token  *oauth2.Token
	client *http.Client
}

type Post struct {
	Author         string                 `json:"author"`
	Commentary     string                 `json:"commentary"`
	Visibility     string                 `json:"visibility"`
	Distribution   map[string]interface{} `json:"distribution"`
	LifecycleState string                 `json:"lifecycleState"`
}

func NewConfig(clientID, clientSecret, redirectURL string) *Config {
	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "w_member_social", "email"},
	}
}

func NewClient(config *Config) *Client {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}

	return &Client{
		config: oauth2Config,
		client: &http.Client{},
	}
}

func (c *Client) GetAuthURL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (c *Client) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	c.token = token
	c.client = c.config.Client(ctx, token)
	return token, nil
}

func (c *Client) SetToken(token *oauth2.Token) {
	c.token = token
	c.client = c.config.Client(context.Background(), token)
}

func (c *Client) GetProfile(ctx context.Context) (map[string]interface{}, error) {
	if c.token == nil {
		return nil, fmt.Errorf("no access token available")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", UserInfoURL, http.NoBody)
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
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
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

func (c *Client) CreatePost(ctx context.Context, text, userID string) error {
	if c.token == nil {
		return fmt.Errorf("no access token available")
	}

	// Create the post payload using the new Posts API format
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

	// Debug: print the post payload
	fmt.Printf("DEBUG: Creating post with author: %s\n", post.Author)
	fmt.Printf("DEBUG: User ID: %s\n", userID)

	jsonData, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", PostsURL, bytes.NewBuffer(jsonData))
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
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
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

func (c *Client) IsAuthenticated() bool {
	return c.token != nil && c.token.Valid()
}
