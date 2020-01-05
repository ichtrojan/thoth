package thoth

import (
	"fmt"
	"os"
	"time"
)

const directory = "logs"

type Config struct {
	directory string
}

func Init() Config {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)

		if err != nil {
			fmt.Println(err)
		}
	}

	path := fmt.Sprintf("%s/error.log", directory)

	var _, err = os.Stat(path)

	if os.IsNotExist(err) {
		file, err := os.Create(path)

		if err != nil {
			fmt.Println(err)
		}

		defer file.Close()
	}

	return Config{directory: path}
}

func (config Config) Log(error error) {
	path := config.directory

	var file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	newError := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), error)

	_, err = fmt.Fprintln(file, newError)

	if err != nil {
		fmt.Println(err)
	}

	err = file.Sync()

	if err != nil {
		fmt.Println(err)
	}

	return
}
