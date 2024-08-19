package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type apiHandler struct {
	r        *chi.Mux
	s3Client *s3.S3
	bucket   string
}

type response struct {
	Urls []string `json:"urls"`
}

func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func NewHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("sa-east-1"),
		Credentials: credentials.NewEnvCredentials(),
	}))

	s3Client := s3.New(sess)
	bucket := "helpme-uploads-bucket"

	apiHandler := apiHandler{
		r,
		s3Client,
		bucket,
	}

	r.Get("/hello", apiHandler.helloWorld)

	r.Post("/upload/local", apiHandler.uploadFile)

	r.Post("/upload/s3", apiHandler.uploadFileToAWSS3)

	return apiHandler
}

func (h apiHandler) helloWorld(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, "hello world")
}

func (h apiHandler) uploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB max memory
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		slog.Error("Failed to parse multipart form")
		return
	}

	// Retrieve the files from the form-data
	files := r.MultipartForm.File["files"]
	if files == nil || len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		slog.Error("No files uploaded")
		return
	}

	// Loop through the files and save each one
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("Failed to open file %s", fileHeader.Filename))
			return
		}
		defer file.Close()

		// Optionally, create a directory to save the file
		dst, err := os.Create(filepath.Join("uploads", fileHeader.Filename))
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("Failed to open dst %s", dst.Name()))
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("Failed to save file %s", fileHeader.Filename))
			return
		}
	}

	sendJSON(w, "files saved successfully")
}

func (h apiHandler) uploadFileToAWSS3(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max memory
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

	// Loop through the files and save each one
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("Failed to open file %s", fileHeader.Filename))
			return
		}
		defer file.Close()

		// Gather file information
		fileName := filepath.Base(fileHeader.Filename)
		normalizedFileName := normalizeFilename(fileName)
		// fileSize := fileHeader.Size
		contentType := fileHeader.Header.Get("Content-Type")
		// uploadTime := time.Now()

		// Upload the file to S3
		_, err = h.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(h.bucket),
			Key:         aws.String(normalizedFileName),
			Body:        file,
			ContentType: aws.String(contentType),
			ACL:         aws.String("private"), // Or "public-read" depending on your requirements
		})
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			slog.Error("Failed to upload file to S3", err)
			return
		}

		// URL or Key for accessing the file in S3
		fileURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", h.bucket, normalizedFileName)
		resp.Urls = append(resp.Urls, fileURL)
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	sendJSON(w, resp, http.StatusCreated)
}
