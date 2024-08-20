package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/FischerRobson/help.me-upload/internal/api"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	handler := api.NewHandler()

	go func() {
		if err := http.ListenAndServe(":8081", handler); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
