package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ichtrojan/thoth"
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

	err = file.Serve("/logs")

	if err != nil {
		fmt.Println(err)
		json.Log(err)
		file.Log(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello, Testing from Thoth")

		if err != nil {
			json.Log(err)
			file.Log(err)
		}
	})

	if err := http.ListenAndServe(":8888", nil); err != nil {
		json.Log(err)
		file.Log(err)
	}
}
