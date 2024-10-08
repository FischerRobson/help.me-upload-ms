package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FischerRobson/help.me-upload/internal/awsProvider"
	"github.com/FischerRobson/help.me-upload/internal/utils"
)

type UploadResponse struct {
	UploadID string   `json:"uploadId"`
	FileURLs []string `json:"fileUrls"`
}

func (h apiHandler) uploadFileWithRabbitMQ(w http.ResponseWriter, r *http.Request) {
	uploadId := r.FormValue("uploadId")
	if uploadId == "" {
		http.Error(w, "Missing uploadId", http.StatusBadRequest)
		return
	}

	err := r.ParseMultipartForm(MAX_UPLOAD_SIZE)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		slog.Error("Failed to parse multipart form", err)
		return
	}

	files := r.MultipartForm.File["files"]
	if files == nil || len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		slog.Error("No files uploaded")
		return
	}

	resp := response{}

	w.WriteHeader(http.StatusOK)
	utils.SendJSON(w, "Files sent to queue", http.StatusCreated)

	uploadToS3 := os.Getenv("ENABLE_AWS_UPLOAD") == "true"

	go func() {
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				slog.Error(fmt.Sprintf("Failed to open file %s", fileHeader.Filename))
				return
			}
			defer file.Close()

			if uploadToS3 {
				fileURL, err := awsProvider.UploadToS3(h.s3Client, w, h.bucket, fileHeader)
				if err != nil {
					continue
				}
				fmt.Print(fileURL)

				resp.Urls = append(resp.Urls, fileURL)
			} else {

				dst, err := os.Create(filepath.Join("uploads", fileHeader.Filename))
				if err != nil {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
					slog.Error(fmt.Sprintf("Failed to open dst %s", dst.Name()))
					return
				}
				defer dst.Close()

				if _, err := io.Copy(dst, file); err != nil {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
					slog.Error(fmt.Sprintf("Failed to save file %s", fileHeader.Filename))
					return
				}
			}
		}

		queueName := os.Getenv("UPLOAD_FILE_QUEUE")
		fmt.Print("Getting queue name")
		if queueName == "" {
			log.Fatal("UPLOAD_FILE_QUEUE environment variable is not set. Exiting application.")
		}

		message := UploadResponse{
			UploadID: uploadId,
			FileURLs: resp.Urls,
		}

		messageBody, err := json.Marshal(message)
		if err != nil {
			fmt.Errorf("Failed to marshal message: %v", err)
			return
		}

		err = h.rabbitMQ.PublishToQueue("fileUploadQueue", messageBody)
		if err != nil {
			log.Printf("Failed to publish message to RabbitMQ: %v", err)
			http.Error(w, "Failed to upload file", http.StatusInternalServerError)
			return
		}
	}()
}
