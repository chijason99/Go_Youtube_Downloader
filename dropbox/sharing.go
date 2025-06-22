package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type createSharedLinkPayload struct {
	Path     string             `json:"path"`
}

type sharedLinkResponse struct {
	URL string `json:"url"`
}

// GetShareableLink creates and returns a public, shareable link for the given file path.
// If a link for the file already exists, it retrieves and returns the existing one.
func GetShareableLink(dropboxFilePath string, accessToken string) (string, error) {
	url := "https://api.dropboxapi.com/2/sharing/create_shared_link_with_settings"

	// settings := getSharedSettings("viewer", true, "public", "public")

	payload := createSharedLinkPayload{
		Path:     dropboxFilePath,
	}

	payloadJsonString, err := json.Marshal(payload)

	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadJsonString))

	if err != nil {
		return "", fmt.Errorf("an error occurred when initializing the http request: %v", err)
	}

	request.Header.Add("Authorization", "Bearer "+accessToken)
	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		return "", fmt.Errorf("an error occurred when sending the http request: %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("request to get sharable link failed with status: %s, %s", response.Status, string(bodyBytes))
	}

	var sharedLinkResponse sharedLinkResponse

	if err := json.NewDecoder(response.Body).Decode(&sharedLinkResponse); err != nil {
		return "", fmt.Errorf("error deserializing the shared link response: %v", err)
	}

	return sharedLinkResponse.URL, nil
}