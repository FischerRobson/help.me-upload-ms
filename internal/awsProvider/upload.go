package awsProvider

import (
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/FischerRobson/help.me-upload/internal/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadToS3(s3Client *s3.S3, w http.ResponseWriter, bucket string,
	f *multipart.FileHeader) (string, error) {
	file, _ := f.Open()

	fileName := filepath.Base(f.Filename)
	normalizedFileName := utils.NormalizeFilename(fileName)
	// fileSize := fileHeader.Size
	contentType := f.Header.Get("Content-Type")
	// uploadTime := time.Now()

	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(normalizedFileName),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         aws.String("private"), // Or "public-read" depending on your requirements
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		slog.Error("Failed to upload file to S3", err)
		return "", err
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, normalizedFileName), nil
}
