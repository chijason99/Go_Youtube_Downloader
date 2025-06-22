package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type DropboxConfig struct {
	RefreshToken string
	AppSecret    string
	AppKey       string
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func LoadDropboxConfig() (*DropboxConfig, error) {
	var dropboxConfig DropboxConfig

	err := godotenv.Load()

	if err != nil {
		return nil, fmt.Errorf("failed to load the .env file: %w", err)
	}

	appSecret := os.Getenv("APP_SECRET")
	appKey := os.Getenv("APP_KEY")
	refreshToken := os.Getenv("REFRESH_TOKEN")

	dropboxConfig = DropboxConfig{
		AppSecret:    appSecret,
		AppKey:       appKey,
		RefreshToken: refreshToken,
	}

	return &dropboxConfig, nil
}

// RefreshAccessToken uses a long-lived refresh token to obtain a new,
// short-lived access token from the Dropbox API.
// It makes use of the dropbox app's key, secret, and the user's refresh token stored inside the .env file and returns a new access token on success.
func GetAccessToken(config *DropboxConfig) (string, error) {
	endpoint := "https://api.dropbox.com/oauth2/token"
	data := url.Values{}

	data.Set("grant_type", "refresh_token")
	data.Set("client_id", config.AppKey)
	data.Set("client_secret", config.AppSecret)
	data.Set("refresh_token", config.RefreshToken)

	request, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))

	if err != nil {
		return "", fmt.Errorf("failed to initialize a http request: %w", err)
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(response.Body)
		
		return "", fmt.Errorf("token request failed with status: %s, %s", response.Status, string(bodyBytes))
	}

	var accessTokenResponse AccessTokenResponse

	if err := json.NewDecoder(response.Body).Decode(&accessTokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return accessTokenResponse.AccessToken, nil
}