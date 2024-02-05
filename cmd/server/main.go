package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jfosburgh/gomes/internal/routes"
)

func main() {

	handleSigTerm()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := routes.NewRouter()

	fmt.Printf("Starting server at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		fmt.Println("ListenAndServe err:", err)
		os.Exit(1)
	}
}

func handleSigTerm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nreceived sigterm, exiting")
		os.Exit(1)
	}()
}
