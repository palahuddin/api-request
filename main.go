package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
	"io"
	"bytes"
	"os"
	"flag"
    "log"
    "time"

    "github.com/gorilla/mux"
 )


type ServiceConf struct {
	Token            string
	Url	string
	Templates string
}

var conf ServiceConf = ServiceConf{}


func GetUser() []byte {

	client := http.Client{}
	req , _ := http.NewRequest("GET", conf.Url, nil)
	req.Header = map[string][]string{
		"Accept": {"application/json"},
        "Content-Type": {"application/json"},
        "Authorization": {"Bearer "+conf.Token},
	}
	resp , _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}


func jsonFilter(username string) string{
    type Config struct {
        Username   string
        Email string
        Id string
    }
	body := GetUser()
    var respbody = []byte(body)
    var config []Config
    err := json.Unmarshal(respbody, &config)
    if err != nil {
        fmt.Println("error:", err)
    }
	
	var userid string
    for _, v := range config {
        if v.Username == username  {
            userid := v.Id
			return userid
        }
    }
	return userid
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	UserId := jsonFilter(r.FormValue("username"))
	PayLoad, _ := json.Marshal(map[string]string{
		"value": r.FormValue("password"),
		"temporary": "false",
	})
	client := http.Client{}
	req , _ := http.NewRequest("PUT", conf.Url+"/"+UserId+"/reset-password", bytes.NewBuffer(PayLoad))
	req.Header = map[string][]string{
		"Accept": {"application/json"},
        "Content-Type": {"application/json"},
        "Authorization": {"Bearer "+conf.Token},
	}
	resp , _ := client.Do(req)
	io.WriteString(w, resp.Status)
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    // A very simple health check.
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    // In the future we could report back on the status of our DB, or our cache
    // (e.g. Redis) by performing a simple PING, and include them in the response.
    io.WriteString(w, `{"alive": true}`)
}

func main() {
	narg := len(os.Args)
	if narg < 2 {
		log.Fatalf("Err::config file not found")
	}

	configFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Err::open config file\n")
	}

	byteValue, _ := ioutil.ReadAll(configFile)
	json.Unmarshal(byteValue, &conf)
	configFile.Close()

	var dir string

    flag.StringVar(&dir, "dir", conf.Templates, "the directory to serve files from. Defaults to the current dir")
    flag.Parse()
    r := mux.NewRouter()
	r.HandleFunc("/health", HealthCheckHandler)
	r.HandleFunc("/reset-password", ChangePassword)

    // This will serve files under http://localhost:8000/<filename>
    r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(dir))))

    srv := &http.Server{
        Handler:      r,
        Addr:         "0.0.0.0:8000",
        // Good practice: enforce timeouts for servers you create!
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }
	fmt.Println("Server Running on Host "+srv.Addr)
    log.Fatal(srv.ListenAndServe())
}


