package main

import (
	"fmt"
	"github.com/ichtrojan/thoth"
	"log"
	"net/http"
)

func main() {
	logger := thoth.InitJson()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello, Testing from Thoth")

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Endpoint served")
	})

	if err := http.ListenAndServe(":8888", nil); err != nil {
		logger.LogJson(err)
	}
}
