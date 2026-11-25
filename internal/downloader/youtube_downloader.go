package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const maxVideoSize = 500 * 1024 * 1024

var ErrInvalidYouTubeURL = errors.New("invalid YouTube video URL")

var videoIDPattern = regexp.MustCompile(
	`^[a-zA-Z0-9_-]{11}$`,
)

type YouTubeDownloader struct {
	storageDir string
}

type DownloadResult struct {
	OriginalName string
	FileName     string
	FilePath     string
	FileSize     int64
	MimeType     string
}

func NewYouTubeDownloader(
	storageDir string,
) *YouTubeDownloader {
	return &YouTubeDownloader{
		storageDir: storageDir,
	}
}

// Validate and normalize the YouTube URL.
func NormalizeYouTubeURL(rawURL string) (string, error) {

	parsed, err := url.Parse(strings.TrimSpace(rawURL))

	if err != nil ||
		parsed.Scheme != "https" ||
		parsed.User != nil ||
		parsed.Port() != "" {

		return "", ErrInvalidYouTubeURL
	}

	var videoID string

	switch strings.ToLower(parsed.Hostname()) {

	case "youtube.com", "www.youtube.com", "m.youtube.com":

		if parsed.Path == "/watch" {

			videoID = parsed.Query().Get("v")

		} else if strings.HasPrefix(parsed.Path, "/shorts/") {

			videoID = strings.TrimPrefix(
				parsed.Path, "/shorts/",
			)
		}

	case "youtu.be":

		videoID = strings.TrimPrefix(parsed.Path, "/")

	default:

		return "", ErrInvalidYouTubeURL
	}

	if !videoIDPattern.MatchString(videoID) {
		return "", ErrInvalidYouTubeURL
	}

	return "https://www.youtube.com/watch?v=" + videoID, nil
}

func (d *YouTubeDownloader) Download(
	ctx context.Context,
	rawURL string,
) (*DownloadResult, error) {

	// Step 1: Validate the URL.
	videoURL, err := NormalizeYouTubeURL(rawURL)
	if err != nil {
		return nil, err
	}

	// Step 2: Verify yt-dlp installation.
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return nil, fmt.Errorf("yt-dlp is not installed: %w", err)
	}

	// Step 3: Create storage directory.
	if err := os.MkdirAll(d.storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Step 4: Create an isolated temporary directory.
	tempDir, err := os.MkdirTemp(d.storageDir, "youtube-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	defer os.RemoveAll(tempDir)

	outputTemplate := filepath.Join(
		tempDir,
		"source.%(ext)s",
	)

	// Step 5: Limit download execution time.
	downloadCtx, cancel := context.WithTimeout(
		ctx,
		10*time.Minute,
	)
	defer cancel()
	denoPath, err := exec.LookPath("deno")
	if err != nil {
		return nil, fmt.Errorf("deno runtime not found: %w", err)
	}
	// Step 6: Prepare yt-dlp arguments.
	args := []string{
		"--ignore-config",
		"--no-playlist",
		"--no-progress",
		"--no-plugin-dirs",

		"--js-runtimes", "deno:" + denoPath,

		"--format",
		"bv[ext=mp4][height<=720]+ba[ext=m4a]/b[ext=mp4][height<=720]",

		"--merge-output-format", "mp4",

		"--max-filesize", "500M",

		"--match-filter", "!is_live & duration <= 3600",

		"--socket-timeout", "30",
		"--retries", "2",

		"--output", outputTemplate,

		"--",
		videoURL,
	}

	// Step 7: Execute downloader.
	cmd := exec.CommandContext(
		downloadCtx,
		"yt-dlp",
		args...,
	)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf(
			"YouTube download failed: %w: %s",
			err,
			string(output),
		)
	}

	if err := downloadCtx.Err(); err != nil {
		return nil, err
	}

	// Step 8: Verify downloaded file.
	tempPath := filepath.Join(tempDir, "source.mp4")

	info, err := os.Stat(tempPath)
	if err != nil {
		return nil, fmt.Errorf(
			"downloaded video not found: %w\nyt-dlp output:\n%s",
			err,
			string(output),
		)
	}

	if !info.Mode().IsRegular() || info.Size() == 0 {
		return nil, fmt.Errorf("downloaded video is invalid")
	}

	if info.Size() > maxVideoSize {
		return nil, fmt.Errorf("downloaded video exceeds 500 MB")
	}

	// Step 9: Move completed file into permanent storage.
	fileName := filepath.Base(tempDir) + ".mp4"

	filePath := filepath.Join(
		d.storageDir,
		fileName,
	)

	if err := os.Rename(tempPath, filePath); err != nil {
		return nil, fmt.Errorf(
			"failed to save downloaded video: %w",
			err,
		)
	}

	parsed, _ := url.Parse(videoURL)
	videoID := parsed.Query().Get("v")

	// Step 10: Return downloaded video details.
	return &DownloadResult{
		OriginalName: "youtube-" + videoID + ".mp4",
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     info.Size(),
		MimeType:     "video/mp4",
	}, nil
}
