package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/FischerRobson/help.me-upload/internal/api"
)

func main() {
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
