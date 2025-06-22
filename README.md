A simple cli application to act as a wrapper of yt-dlp, also allows to upload the video/audio downloaded to the dedicated dropbox application

Flags:
- mp3: a boolean flag to indicate you only want the audio
- dropbox: a boolean flag to indicate you want to upload the downloaded file to your dropbox app

```bash
go run main.go -url https://youtu.be/dQw4w9WgXcQ -mp3 -dropbox
```

Prerequisites:
You would need to have installed FFMPEG and YT-DLP as this script is essentially a wrapper around these services.

Go to config.json file, change the path_to_yt_dlp_directory to your own directory that contains the yt-dlp exe file

Dropbox:
For using the dropbox functionality, create a .env file in the same directory as main.go, with the following properties:

REFRESH_TOKEN = <the-refresh-token-of-your-dropbox-app>
APP_KEY = <the-client_key-of-your-dropbox-app>
APP_SECRET = <the-client-secret-of-your-dropbox-app>
