package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ichtrojan/thoth"
)

func main() {
	logger := thoth.Init("log")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello, Testing from Thoth")

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Endpoint served")
	})

	if err := http.ListenAndServe(":8888", nil); err != nil {
		logger.Log(err)
	}
}
