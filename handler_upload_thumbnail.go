package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// maxMemory represents 10MB of memory by bitshifting the number 10 20 times.
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil { // Respond with a 400 code if the form file can not be parsed.
		respondWithError(w, http.StatusBadRequest, "unable to parse form file", err)
		return
	}
	defer file.Close()

	// Get media type from `Content-Type` header.
	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil { // Respond with a 500 code if media type can not be extracted from the Content-Type header.
		respondWithError(w, http.StatusInternalServerError, "invalid Content-Type header", err)
		return
	}
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		// Respond with a 400 code if media type is not a jpeg or png.
		respondWithError(w, http.StatusBadRequest, "invalid file type: thumbnail should be a jpeg or png", nil)
		return
	}

	// Get the asset path string and the actual asset path on disk.
	assetPath := getAssetPath(videoID, mediaType)
	assetDiskPath := cfg.getAssetDiskPath(assetPath)

	// Create new file on disk.
	newFile, err := os.Create(assetDiskPath)
	if err != nil { // Respond with a 500 code if file can not be created.
		respondWithError(w, http.StatusInternalServerError, "unable to create file on server", err)
	}
	defer newFile.Close()
	// Copy contents from multipart file to new file.
	if _, err = io.Copy(newFile, file); err != nil { // Respond with a 500 code if the file can not be copied to new file on disk.
		respondWithError(w, http.StatusInternalServerError, "unable to copy file data to new file", err)
		return
	}

	// Get the video metadata from database.
	video, err := cfg.db.GetVideo(videoID)
	if err != nil { // Respond with a 404 code if video can not be retrieved from database.
		respondWithError(w, http.StatusNotFound, "unable to retrieve video metadata", err)
		return
	}
	if userID != video.UserID { // Respond with a 401 code if user is not the video owner.
		respondWithError(w, http.StatusUnauthorized, "user is not authorized to update this video", err)
		return
	}

	// Generate URL to asset.
	dataURL := cfg.getAssetURL(assetPath)

	// Set the URL in the metadata.
	video.ThumbnailURL = &dataURL

	// Update the database record for the video to include the thumbnail information.
	err = cfg.db.UpdateVideo(video)
	if err != nil { // Respond with a 500 code if the video metadata can not be updated.
		respondWithError(w, http.StatusInternalServerError, "unable to update video metadata in database", err)
		return
	}
	// Respond with a 200 code and the metaData struct indicating success.
	respondWithJSON(w, http.StatusOK, video)
}
