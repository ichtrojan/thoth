package thoth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const directory = "logs"

type Config struct {
	directory string
	filetype  string
}

func Init(filetype string) (Config, error) {
	var config Config

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)

		if err != nil {
			return config, err
		}
	}

	var filename string

	switch filetype {
	case "log":
		filename = "error.log"
	case "json":
		filename = "error.json"
	default:
		return config, errors.New("adapter not defined")
	}

	path := fmt.Sprintf("%s/%s", directory, filename)

	var _, err = os.Stat(path)

	if os.IsNotExist(err) {
		file, err := os.Create(path)

		if err != nil {
			fmt.Println(err)
		}

		defer file.Close()
	}

	return Config{directory: path, filetype: filetype}, nil
}

func (config Config) Log(error error) {
	switch config.filetype {
	case "log":
		_ = config.logFile(error)
	case "json":
		_ = config.logJson(error)
	default:
		return
	}
}

func (config Config) logFile(error error) error {
	path := config.directory

	var file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	newError := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), error.Error())

	_, err = fmt.Fprintln(file, newError)

	if err != nil {
		return err
	}

	err = file.Sync()

	if err != nil {
		return err
	}

	return nil
}

func (config Config) logJson(error error) error {
	path := config.directory

	var file, err = ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	var jsonData []map[string]interface{}

	_ = json.Unmarshal(file, &jsonData)

	newError := map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"error":     error.Error(),
	}

	jsonData = append(jsonData, newError)

	jsonString, err := json.Marshal(jsonData)

	if err != nil {
		return err
	}

	_ = ioutil.WriteFile(path, jsonString, os.ModePerm)

	return nil
}
