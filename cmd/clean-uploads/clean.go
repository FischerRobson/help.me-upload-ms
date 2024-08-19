package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func main() {
	uploadsDir := "uploads"

	dir, err := os.Open(uploadsDir)
	if err != nil {
		slog.Error("Failed to open dir", err)
	}
	defer dir.Close()

	files, err := dir.Readdirnames(-1)
	if err != nil {
		slog.Error("Failed to read files from dir", err)
	}

	for _, file := range files {
		if file == ".gitkeep" {
			continue
		}

		filePath := filepath.Join(uploadsDir, file)
		err := os.Remove(filePath)
		if err != nil {
			slog.Error("Failed to delete file", file, "error", err)
		}
		slog.Info("File removed", filePath)
	}

	fmt.Println("Uploads directory cleared successfully!")
}
