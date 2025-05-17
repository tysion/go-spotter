package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type HelloResponse struct {
	Name string `json:"name"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	var response HelloResponse

	response.Name = r.URL.Query().Get("name")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/hello", helloHandler)

	fmt.Println("Starting server at :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}