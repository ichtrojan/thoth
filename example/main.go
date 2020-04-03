package main

import (
	"fmt"
	"github.com/ichtrojan/thoth"
	"log"
	"net/http"
)

func main() {
	json, err := thoth.Init("json")

	if err != nil {
		log.Fatal(err)
	}

	file, err := thoth.Init("log")

	if err != nil {
		log.Fatal(err)
	}

	if err := file.Serve("/logs", "12345"); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello, Testing from Thoth")
	})

	if err := http.ListenAndServe(":8888", nil); err != nil {
		file.Log(err)
		json.Log(err)
	}
}
