package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FischerRobson/help.me-upload/internal/awsProvider"
	"github.com/FischerRobson/help.me-upload/internal/utils"
)

const MAX_UPLOAD_SIZE = 10 << 20 // 10MB max memory

type response struct {
	Urls []string `json:"urls"`
}

func (h apiHandler) uploadFile(w http.ResponseWriter, r *http.Request) {
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

	uploadToS3 := os.Getenv("UPLOAD_TO_S3") == "true"

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

	w.WriteHeader(http.StatusOK)
	utils.SendJSON(w, "Files sent to queue", http.StatusCreated)
}
