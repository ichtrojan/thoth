package thoth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

const directory = "logs"

const jsonDashboardView = "./views/dashboardJson.gohtml"
const logsDashboardView = "./views/dashboardLogs.gohtml"

type Config struct {
	directory string
	filetype  string
	key       string
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

func (config Config) Serve(endpoint string, key string) error {

	filename = config.directory

	config.key = key

	authUrl := "/auth"

	http.HandleFunc(endpoint, config.serveHome)
	http.HandleFunc(authUrl, config.checkAuth)
	http.HandleFunc("/ws", serveWs)

	return nil
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

	newError := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), error.Error())

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

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 1 * time.Second
)

var (
	homeTempl, _ = template.ParseFiles(jsonDashboardView)
	logsTempl, _ = template.ParseFiles(logsDashboardView)
	filename     string
	upgrader     = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func (config Config) checkAuth(w http.ResponseWriter, r *http.Request) {
	err := r.Header["Key"][0]

	if err == config.key {
		response := map[string]string{"status": "success"}

		_ = json.NewEncoder(w).Encode(&response)
	} else {
		response := map[string]string{"status": "failed"}

		_ = json.NewEncoder(w).Encode(&response)
	}
}

func readFileIfModified(lastMod time.Time) ([]byte, time.Time, error) {
	fileInfo, err := os.Stat(filename)

	if err != nil {
		return nil, lastMod, err
	}

	if !fileInfo.ModTime().After(lastMod) {
		return nil, lastMod, nil
	}

	fileData, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, fileInfo.ModTime(), err
	}

	return fileData, fileInfo.ModTime(), nil
}

func reader(ws *websocket.Conn) {
	defer ws.Close()

	ws.SetReadLimit(512)

	_ = ws.SetReadDeadline(time.Now().Add(pongWait))

	ws.SetPongHandler(func(string) error {
		_ = ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := ws.ReadMessage()

		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn, lastMod time.Time) {
	lastError := ""

	pingTicker := time.NewTicker(pingPeriod)

	fileTicker := time.NewTicker(filePeriod)

	defer func() {
		pingTicker.Stop()
		fileTicker.Stop()
		_ = ws.Close()
	}()

	for {
		select {
		case <-fileTicker.C:
			var fileData []byte
			var err error

			fileData, lastMod, err = readFileIfModified(lastMod)

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					fileData = []byte(lastError)
				}
			} else {
				lastError = ""
			}

			if fileData != nil {
				_ = ws.SetWriteDeadline(time.Now().Add(writeWait))

				if err := ws.WriteMessage(websocket.TextMessage, fileData); err != nil {
					return
				}
			}
		case <-pingTicker.C:
			_ = ws.SetWriteDeadline(time.Now().Add(writeWait))

			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	var lastMod time.Time

	if n, err := strconv.ParseInt(r.FormValue("lastMod"), 16, 64); err == nil {
		lastMod = time.Unix(0, n)
	}

	go writer(ws, lastMod)

	reader(ws)
}

func (config Config) serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fileData, lastMod, err := readFileIfModified(time.Time{})

	if err != nil {
		fileData = []byte(err.Error())
		lastMod = time.Unix(0, 0)
	}

	returnData := struct {
		Host    string
		Data    string
		LastMod string
	}{
		r.Host,
		string(fileData),
		strconv.FormatInt(lastMod.UnixNano(), 16),
	}

	switch config.filetype {
	case "log":
		_ = logsTempl.Execute(w, &returnData)
	case "json":
		_ = homeTempl.Execute(w, &returnData)
	default:
		return
	}
}
