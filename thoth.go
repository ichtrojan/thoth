package thoth

import (
	"fmt"
	"os"
	"time"
	"encoding/json"
	"io/ioutil"
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

func InitJson(params ...string) Config {
	path := fmt.Sprintf("%s/error.json", directory)

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

	newError := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), error.Error())

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

func (config Config) LogJson(error error) {
	path := config.directory

	var file, err = ioutil.ReadFile(path)

	if err != nil {
		fmt.Println(err)
	}
	jsonData := []map[string]interface{}{}
	json.Unmarshal(file, &jsonData)

	newError := map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"error": error.Error(),
	}

	jsonData = append(jsonData, newError)
	jsonString, _ := json.Marshal(jsonData)

	ioutil.WriteFile(path, jsonString, os.ModePerm)
	return
}
