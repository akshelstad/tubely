package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
)

// getVideoAspectRatio returns the aspect ratio of a video at the given path by calculating the ratio based on the width and heigth of the video.
func getVideoAspectRatio(filePath string) (string, error) {
	// Create a new ffprobe command to retrieve the video metadata.
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_streams",
		filePath,
	)

	// Create a new buffer to store command output.
	var stdout bytes.Buffer
	cmd.Stdout = &stdout // Set the buffer as the stdout destination from the command.

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffprobe error: %v", err)
	}

	// Unmarshal data from the buffer to a videoMetadata json struct.
	var output struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return "", fmt.Errorf("unable to parse ffprobe output: %v", err)
	}
	// Check that the output struct contains data.
	if len(output.Streams) == 0 {
		return "", errors.New("no video streams found")
	}

	// Extract the width and height from the json struct.
	width := output.Streams[0].Width
	height := output.Streams[0].Height

	// Calculate the aspect ratio using the helper function
	aspectRatio := calculateAspectRatio(width, height)

	return aspectRatio, nil
}

// calculateAspectRatio is a helper function that calculates the ratio (in float64) from a given width and height
// and returns the closest match (within 0.01) from a map of float64 values to aspect ratio strings.
// If a closest match isn't found, it formats the width and height into an aspect ratio string.
func calculateAspectRatio(width, height int) string {
	arMap := map[float64]string{
		16.0 / 9.0: "16:9",
		9.0 / 16.0: "9:16",
		4.0 / 3.0:  "4:3",
		3.0 / 4.0:  "3:4",
		1.0:        "1:1",
	}

	ratio := float64(width) / float64(height)

	for key, label := range arMap {
		if (math.Abs(ratio - key)) < 0.01 {
			return label
		}
	}

	return fmt.Sprintf("%d:%d", width, height)
}

// processVideoForFastStart moves the moov atom to the start of the video file and returns the file path to the processed video.
func processVideoForFastStart(filePath string) (string, error) {
	// Create output file path.
	filePathOut := fmt.Sprintf("%s.processing", filePath)

	// Create a new ffmpeg command to move the moov atom to the start of the file.
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", filePathOut)
	// Create a new buffer for the command to write any errors to.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error processing video: %s, %v", stderr.String(), err)
	}

	// Retrieve the file info to check if it is successfully processed.
	fileInfo, err := os.Stat(filePathOut)
	if err != nil {
		return "", fmt.Errorf("unable to stat processed file: %v", err)
	}
	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("processed file is empty")
	}

	return filePathOut, nil
}
