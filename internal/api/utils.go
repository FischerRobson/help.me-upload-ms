package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

func sendJSON(w http.ResponseWriter, rawData any, statusCode ...int) {
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	data, _ := json.Marshal(rawData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func normalizeFilename(filename string) string {
	// Replace spaces with hyphens and remove special characters
	re := regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)
	normalizedFilename := re.ReplaceAllString(filename, "-")
	normalizedFilename = strings.ToLower(normalizedFilename)
	return normalizedFilename
}
