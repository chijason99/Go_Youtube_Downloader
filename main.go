package main

import (
	"bytes"
	"chijason99/go_youtube_downloader/dropbox"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Config is the environment settings that is located in the config.json file
type Config struct {
	PathToYtDlpDirectory  string `json:"path_to_yt_dlp_directory"`
	AudioFormat           string `json:"audio_format"`
	DefaultOutputTemplate string `json:"default_output_template"`
}

// YtdlpInfo is a holder of the metadata of the video that the user is trying to download
type YtdlpInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// UserOptions is a struct that holds the flags passed in by the user when running the script
type UserOptions struct {
	URL             string
	AudioOnly       bool
	UploadToDropbox bool
}

func main() {
	// A command line tool that allows downloading video from YouTube with the help of yt-dlp
	config := loadConfig()

	// // Parsing the flags
	options := setUserOptions()

	ytDlpPath := fmt.Sprint(config.PathToYtDlpDirectory, "/yt-dlp")

	videoInfo, err := fetchVideoInfo(ytDlpPath, options.URL)

	if err != nil {
		log.Fatalf("failed to fetch video info: %v", err)
	}

	// The name of the downloaded file. Will be used for uploading to dropbox
	localFileName := buildOutputFilename(config.DefaultOutputTemplate, videoInfo, options.AudioOnly, config)

	if err := downloadFile(ytDlpPath, options.URL, options.AudioOnly, *config, localFileName); err != nil {
		log.Fatalf("yt-dlp command failed: %v", err)
	}

	if options.UploadToDropbox {
		// Get the access token for the dropbox app
		dropboxConfig, err := dropbox.LoadDropboxConfig()

		if err != nil {
			log.Fatalf("an error occurred when loading in the dropbox config: %s", err)
		}

		accessToken, err := dropbox.GetAccessToken(dropboxConfig)

		if err != nil {
			log.Fatalf("an error occurred when trying to get the dropbox access token: %s", err)
		}

		dropboxFilePath, err := dropbox.UploadFile(localFileName, accessToken)

		if err != nil {
			log.Fatalf("Failed to upload to Dropbox: %v", err)
		}

		log.Printf("File successfully uploaded to Dropbox: %s\n", dropboxFilePath)

		fmt.Printf("Checking if file ID '%s' exists on Dropbox...\n", dropboxFilePath)
		exists, err := dropbox.CheckFileExists(dropboxFilePath, accessToken)
		if err != nil {
			log.Fatalf("Error during existence check: %v", err)
		}

		if !exists {
			log.Fatalf("FATAL: File does not exist on Dropbox immediately after upload.")
		}

		fmt.Println("SUCCESS: File found on Dropbox. Proceeding to get share link.")

		sharableLink, err := dropbox.GetShareableLink(dropboxFilePath, accessToken)

		if err != nil {
			log.Fatalf("an error occurred when trying to get the sharable link for the uploaded file: %s", err)
		}

		log.Printf("Sharable link retrieved: %s", sharableLink)
	}
}

func loadConfig() *Config {
	// Open the config.json file
	jsonFile, err := os.Open("PATH_TO_CONFIG_FILE/config.json") // Replace with the actual path to your config.json file

	if err != nil {
		log.Fatal(err)
	}

	defer jsonFile.Close()

	// read the json file into bytes array
	byteValue, _ := io.ReadAll(jsonFile)

	var config Config

	json.Unmarshal(byteValue, &config)

	return &config
}

func setUserOptions() *UserOptions {
	url := flag.String("url", "", "The YouTube URL to download.")
	audioOnly := flag.Bool("mp3", false, "Download audio only.")
	uploadToDropbox := flag.Bool("dropbox", false, "Upload the final file to Dropbox.")

	flag.Parse()

	if *url == "" {
		log.Fatal("Error: The -url flag is required.")
	}

	// Create the struct and populate it with the *values* from the flags.
	return &UserOptions{
		URL:             *url,
		AudioOnly:       *audioOnly,
		UploadToDropbox: *uploadToDropbox,
	}
}

func fetchVideoInfo(ytDlpPath string, url string) (*YtdlpInfo, error) {
	fmt.Println("Fetching video metadata...")
	var videoInfo YtdlpInfo

	cmd := exec.Command(ytDlpPath, "--print-json", "--simulate", url)

	var infoBuffer bytes.Buffer
	cmd.Stdout = &infoBuffer

	// We can ignore stderr here as simulate rarely fails verbosely
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}

	if err := json.Unmarshal(infoBuffer.Bytes(), &videoInfo); err != nil {
		return nil, fmt.Errorf("failed to parse video metadata JSON: %w", err)
	}
	return &videoInfo, nil
}

func buildOutputFilename(template string, info *YtdlpInfo, audioOnly bool, config *Config) string {
	outputFile := template

	// Sanitize title to remove characters that are invalid in file paths
	sanitizedTitle := strings.ReplaceAll(info.Title, ":", "_")
	sanitizedTitle = strings.ReplaceAll(sanitizedTitle, "?", "_")
	sanitizedTitle = strings.ReplaceAll(sanitizedTitle, " ", "_")

	outputFile = strings.ReplaceAll(outputFile, "%(title)s", sanitizedTitle)
	outputFile = strings.ReplaceAll(outputFile, "%(id)s", info.ID)
	outputFile = strings.ReplaceAll(outputFile, "/", "\\")
	outputFile = strings.ReplaceAll(outputFile, " ", "_")

	if audioOnly {
		outputFile = strings.ReplaceAll(outputFile, "%(ext)s", config.AudioFormat)
	} else {
		outputFile = strings.ReplaceAll(outputFile, "%(ext)s", "mp4")
	}

	return outputFile
}

func downloadFile(ytDlpPath string, url string, audioOnly bool, config Config, fileName string) error {
	// Creating a slice of strings to store the arguments
	var args []string

	if audioOnly {
		args = append(args, "-f", "bestaudio", "-x", "--audio-format", config.AudioFormat)
	}

	args = append(args, "-o", fileName)

	// Always append the URL last
	args = append(args, url)

	fmt.Printf("Download started...")

	cmd := exec.Command(ytDlpPath, args...)

	// Connecting the output from the cmd to the running console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
