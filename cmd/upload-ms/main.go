package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/FischerRobson/help.me-upload/internal/api"
	"github.com/FischerRobson/help.me-upload/internal/rabbitmq"
	"github.com/joho/godotenv"
)

func main() {

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	rabbitMQService, err := rabbitmq.NewRabbitMQService()
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rabbitMQService.Close()

	handler := api.NewHandler(rabbitMQService)

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
