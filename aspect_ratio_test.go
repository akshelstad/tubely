package main

import (
	"log"
	"testing"
)

func TestCalculateAspectRatio(t *testing.T) {
	tests := []struct {
		name            string
		width           int
		height          int
		wantAspectRatio string
	}{
		{
			name:            "1280x720",
			width:           1280,
			height:          720,
			wantAspectRatio: "16:9",
		},
		{
			name:            "1920x1080",
			width:           1920,
			height:          1080,
			wantAspectRatio: "16:9",
		},
		{
			name:            "1024x768",
			width:           1024,
			height:          768,
			wantAspectRatio: "4:3",
		},
		{
			name:            "3840x2160",
			width:           3840,
			height:          2160,
			wantAspectRatio: "16:9",
		},
		{
			name:            "640x480",
			width:           640,
			height:          480,
			wantAspectRatio: "4:3",
		},
		{
			name:            "608x1080",
			width:           608,
			height:          1080,
			wantAspectRatio: "9:16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arWidth, arHeight := calculateAspectRatio(tt.width, tt.height)
			// gotAspectRatio := fmt.Sprintf("%d:%d", arWidth, arHeight)
			gotAspectRatio := calculateAspectRatio(tt.width, tt.height)
			if gotAspectRatio != tt.wantAspectRatio {
				t.Errorf("calculateAspectRatio() gotAspectRatio = %s, want = %s", gotAspectRatio, tt.wantAspectRatio)
				return
			}
			log.Printf("calculateAspectRatio() gotAspectRatio = %s, want = %s - OK", gotAspectRatio, tt.wantAspectRatio)
		})
	}
}

func TestGetVideoAspectRatio(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		wantAspectRatio string
	}{
		{
			name:            "sample horizontal video",
			path:            "./samples/boots-video-horizontal.mp4",
			wantAspectRatio: "16:9",
		},
		{
			name:            "sample vertical video",
			path:            "./samples/boots-video-vertical.mp4",
			wantAspectRatio: "9:16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAspectRatio, err := getVideoAspectRatio(tt.path)
			if err != nil {
				t.Errorf("getVideoAspectRatio() error = %v", err)
				return
			}
			if gotAspectRatio != tt.wantAspectRatio {
				t.Errorf("getVideoAspectRatio() gotAspectRatio = %s, want = %s", gotAspectRatio, tt.wantAspectRatio)
				return
			}
		})
	}
}
