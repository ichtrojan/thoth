package main

import (
	"fmt"
	"log"
	"net/http"
	"errors"

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

	err = file.Serve("/logs","12345")

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

		er := errors.New("Starting your server")

		file.Log(er)
		json.Log(er)

		json.Log(err)
		file.Log(err)
	}
}
