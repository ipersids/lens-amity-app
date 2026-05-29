package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got /api/health request")
	io.WriteString(w, "I'm okay.")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getHealth)

	err := http.ListenAndServe(":8080", mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
