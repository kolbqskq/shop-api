package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"shop-api/internal/app"
	"syscall"
	"time"
)

func main() {
	server := app.Run()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown error")
	}

}
