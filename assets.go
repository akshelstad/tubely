package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

// getAssetPath generates a path for an asset based on the videoID and media type.
// It uses the helper function mediaTypeToExt to retrieve the file extension.
func getAssetPath(videoID uuid.UUID, mediaType string) string {
	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", videoID, ext)
}

// getAssetDiskPath generates the path on disk for an asset. It uses the asset path root defined in the configuration.
func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

// getAssetURL generates a URL for the asset.
func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

// mediaTypeToExt is a helper function that extracts the file extension from the media type string.
func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}
