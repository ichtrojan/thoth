package main

import (
	"fmt"
	"net/http"

	"github.com/ichtrojan/thoth"
)

func main() {
	json, err := thoth.Init("json")

	if err != nil {
		fmt.Println(err)
	}
	

	file, err := thoth.Init("log")

	if err != nil {
		fmt.Println(err)
	}

	err = file.Serve("/mylogs")

	if err != nil {
		fmt.Println(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello, Testing from Thoth")

		if err != nil {
			json.Log(err)
		}

		fmt.Println("Endpoint served")
	})

	if err := http.ListenAndServe(":8888", nil); err != nil {
		json.Log(err)
		file.Log(err)
		
	}
}
