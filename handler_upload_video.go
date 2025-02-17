package main

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const maxUpload = 1 << 30 // Max upload size of 1GB

	// Use http.MaxBytesReader to limit size of uploads on request body.
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	// Get string version of video ID from request path & parse to UUID.
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil { // Respond with a 400 code if ID can not be parsed to UUID.
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	// Get the JWT token from the header.
	token, err := auth.GetBearerToken(r.Header)
	if err != nil { // Respond with a 401 code if JWT token can not be retrieved.
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil { // Respond with a 401 code if JWT token can not be validated.
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
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

	// Parse uploaded video file from form data.
	file, header, err := r.FormFile("video")
	if err != nil { // Respond with a 400 code if teh form file can not be parsed.
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
	if mediaType != "video/mp4" { // Respond with a 400 code if media type is not an mp4.
		respondWithError(w, http.StatusBadRequest, "invalid file type: video should be an mp4", nil)
		return
	}

	// Create a temporary file on disk for the upload. Delete after stored in s3.
	tmpFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil { // Respond with a 500 code if the temporary file can not be created.
		respondWithError(w, http.StatusInternalServerError, "unable to create temp file on disk", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy contents from multipart file to new file.
	if _, err = io.Copy(tmpFile, file); err != nil { // Respond with a 500 code if the file can not be copied to new file on disk.
		respondWithError(w, http.StatusInternalServerError, "unable to copy file data to new file", err)
		return
	}

	// Reset pointer of temp file to beginning of file.
	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil { // Respond with a 500 code if the file pointer can not be set to the beginning of the file.
		respondWithError(w, http.StatusInternalServerError, "unable to reset pointer to beginning of file", err)
		return
	}

	// // Get the absolute path of the temporary file.
	// tmpPath, err := filepath.Abs(tmpFile.Name())
	// if err != nil { // Respond with a 500 code if the file path for the temp video file can not be retrieved.
	// 	respondWithError(w, http.StatusInternalServerError, "unable to retrieve absolute file path of temp file", err)
	// 	return
	// }

	// Get the aspect ratio of the video.
	aspectRatio, err := getVideoAspectRatio(tmpFile.Name())
	if err != nil { // Respond with a 500 code if the aspect ratio for the video file can not be retrieved.
		respondWithError(w, http.StatusInternalServerError, "unable to retrieve aspect ratio from video", err)
		return
	}

	// Create and set orientation variable based on aspectRatio.
	var orientation string

	switch aspectRatio {
	case "16:9":
		orientation = "landscape"
	case "9:16":
		orientation = "portrait"
	default:
		orientation = "other"
	}

	// Generate the video key: <random-32-byte-hex>.ext.
	videoKey := fmt.Sprintf("%s/%s", orientation, getAssetPath(mediaType))

	// Process the video for fast start encoding.
	processedVideoPath, err := processVideoForFastStart(tmpFile.Name())
	if err != nil { // Respond with a 500 code if the video can not be processed for fast start encoding.
		respondWithError(w, http.StatusInternalServerError, "unable to process video", err)
		return
	}
	defer os.Remove(processedVideoPath)

	// Open the processed video file. Delete when done uploading.
	processedFile, err := os.Open(processedVideoPath)
	if err != nil { // Respond with a 500 code if the processed video can not be opened.
		respondWithError(w, http.StatusInternalServerError, "unable to open processed video file", err)
		return
	}
	defer processedFile.Close()

	// Put video object into s3 bucket.
	_, err = cfg.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.s3Bucket),
		Key:         aws.String(videoKey),
		Body:        processedFile, // Upload the processed video instead of original temp file.
		ContentType: aws.String(mediaType),
	})
	if err != nil { // Respond with a 500 code if the video can not be uploaded to s3 bucket.
		respondWithError(w, http.StatusInternalServerError, "unable to upload file to s3 bucket", err)
		return
	}

	// Generate AWS video URL & update video metadata in database.
	videoURL := cfg.getObjectURL(videoKey)
	video.VideoURL = &videoURL
	err = cfg.db.UpdateVideo(video)
	if err != nil { // Respond with a 500 code if the video metadata can not be updated.
		respondWithError(w, http.StatusInternalServerError, "unable to update video metadata in database", err)
		return
	}

	// Respond with a 200 code and the metaData struct indicating success.
	respondWithJSON(w, http.StatusOK, video)
}
