package api

import (
	"log"
	"net/http"
	"os"

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

	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		log.Fatal("S3_BUCKET_NAME environment variable is not set")
	}

	s3Client := s3.New(sess)

	apiHandler := apiHandler{
		r,
		s3Client,
		bucket,
	}

	r.With(JWTMiddleware).Post("/upload", apiHandler.uploadFile)

	return apiHandler
}
