// Package dropbox provides functions for interacting with the Dropbox API.
// It handles getting the access token for the dropbox app, file uploading, and acquiring a sharable link for the uploaded file.
package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type DropboxUploadArgs struct {
	Path       string `json:"path"`
    Mode       string `json:"mode"`
    AutoRename bool   `json:"autorename"`
    Mute       bool   `json:"mute"`
}

type FileUploadResponse struct {
	ID string `json:"id"`
	Path string `json:"path_lower"`
}

type getMetadataRequest struct {
    Path                           string `json:"path"`
    IncludeDeleted                 bool   `json:"include_deleted"`
    IncludeHasExplicitSharedMembers bool   `json:"include_has_explicit_shared_members"`
    IncludeMediaInfo               bool   `json:"include_media_info"`
}

// UploadFile uploads a local file to a specified path in Dropbox.
// It requires the path to the local file, and returns the path of the uploaded file in the dropbox app
// Returns the id of the uploaded file
func UploadFile(localFilePath string, accessToken string) (string, error) {
	DROPBOX_UPLOAD_PATH := "https://content.dropboxapi.com/2/files/upload"

	file, err := os.Open(localFilePath)

	if err != nil{
		return "", err
	}

	defer file.Close()

	fileName := filepath.Base(localFilePath)

	request, err := http.NewRequest("POST", DROPBOX_UPLOAD_PATH, file)

	if err != nil {
		return "", fmt.Errorf("an error occurred while initializing the http request: %w", err)
	}

	uploadArgs, err := getUploadArgs(fileName)

	if err != nil{
		return "", fmt.Errorf("an error occurred while getting the dropbox api upload args: %w", err)
	}

	request.Header.Add("Content-Type", "application/octet-stream")
	request.Header.Add("Authorization", "Bearer " + accessToken)
	request.Header.Add("Dropbox-API-Arg", uploadArgs)

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return "", fmt.Errorf("an error occurred while sending the http request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("file upload request failed with status: %s", response.Status)
	}

	defer response.Body.Close()

	var fileUploadResponse FileUploadResponse;

	if err := json.NewDecoder(response.Body).Decode(&fileUploadResponse); err != nil{
		return "", fmt.Errorf("an error occurred while deserializing the file upload response from Dropbox: %w", err)
	}

	return fileUploadResponse.ID, nil
}

func getUploadArgs(fileName string) (string, error){
	uploadArgs := DropboxUploadArgs{
		Path: "/downloads/" + fileName,
		AutoRename: false,
		Mute: false,
		Mode: "add",
	}

	uploadArgsJson, err := json.Marshal(uploadArgs)

	if err != nil{
		return "", err
	}

	return string(uploadArgsJson), nil
}

func CheckFileExists(identifier string, accessToken string) (bool, error) {
    endpoint := "https://api.dropboxapi.com/2/files/get_metadata"

    requestBody := getMetadataRequest{
        Path:                           identifier,
        IncludeDeleted:                 false,
        IncludeHasExplicitSharedMembers: false,
        IncludeMediaInfo:               false,
    }

    jsonData, err := json.Marshal(requestBody)
    if err != nil {
        return false, fmt.Errorf("failed to marshal metadata request: %w", err)
    }

    req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
    if err != nil {
        return false, fmt.Errorf("failed to create metadata request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Content-Type", "application/json")

	apiClient := &http.Client{}

    // Assuming you have a shared apiClient as recommended
    resp, err := apiClient.Do(req)
    if err != nil {
        return false, fmt.Errorf("failed to send metadata request: %w", err)
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case http.StatusOK:
        // 200 OK means the file was found.
        return true, nil
    case http.StatusConflict:
        // For this endpoint, a 409 Conflict specifically means the path was not found.
        return false, nil
    default:
        // Any other status is an unexpected error.
        bodyBytes, _ := io.ReadAll(resp.Body)
        return false, fmt.Errorf("unexpected error checking metadata, status %s: %s", resp.Status, string(bodyBytes))
    }
}