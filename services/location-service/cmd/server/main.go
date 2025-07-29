package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	fmt.Println("Location Service starting on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Location Service is healthy!"))
}
