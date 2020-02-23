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


func (config Config) Serve(endpoint string) error{

	filename = config.directory

	http.HandleFunc(endpoint, serveHome)
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

	homeTempl = template.Must(template.New("").Parse(homeHTML))
	filename  string
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func readFileIfModified(lastMod time.Time) ([]byte, time.Time, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, lastMod, err
	}
	if !fi.ModTime().After(lastMod) {
		return nil, lastMod, nil
	}
	p, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fi.ModTime(), err
	}
	return p, fi.ModTime(), nil
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
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
		ws.Close()
	}()
	for {
		select {
		case <-fileTicker.C:
			var p []byte
			var err error

			p, lastMod, err = readFileIfModified(lastMod)

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					p = []byte(lastError)
				}
			} else {
				lastError = ""
			}

			if p != nil {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
					return
				}
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
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

func serveHome(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	p, lastMod, err := readFileIfModified(time.Time{})

	if err != nil {
		p = []byte(err.Error())
		lastMod = time.Unix(0, 0)
	}
	var v = struct {
		Host    string
		Data    string
		LastMod string
	}{
		r.Host,
		string(p),
		strconv.FormatInt(lastMod.UnixNano(), 16),
	}
	homeTempl.Execute(w, &v)
}


const homeHTML = `<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">

    <title>Thoth Dashboard</title>
  </head>
  <body>
    <div class="wrap">
        <p><h3>Thoth logs</h3></p>
        <div class="tool-bar">
            
        </div>
        <div class="card">
            <div class="card-body ">
                <pre id="fileData" class="white-text">{{.Data}}</pre>
            </div>
        </div>
    </div>

    <!-- Optional JavaScript -->
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
	<script type="text/javascript">
		var passw = prompt("enter secure key");
		var real = "1234"
		if(passw == real){
			(function() {
				// var seckKey = prompt("enter secure key");
				// // prolly use this for authentication
				// console.log(seckKey)
				var data = document.getElementById("fileData");
				data.style.color = "white"
				var conn = new WebSocket("ws://{{.Host}}/ws?lastMod={{.LastMod}}");
				conn.onclose = function(evt) {
					data.textContent = 'Connection closed';
				}
				conn.onmessage = function(evt) {
					console.log('file updated');
					console.log(evt)
					data.textContent = evt.data;
				}

			})();
		}else{
			alert("invalid login details");
			var data = document.getElementById("fileData");
			data.textContent = 'Invalid security key';
		}
        
    </script>
    
    <style>
    
    .wrap{
        margin:auto;
        width:90%;
    }
    /* .card{
        box-shadow: 2px 2px 5px 2px rgba(53, 196, 60, 0.3);
     }  */
    .card-body{
        padding:20px;
        background-color:black; 
        color:white;
    } 
    </style>
    <script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>
  </body>
</html>


`
