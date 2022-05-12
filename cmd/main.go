package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type APIHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type APIDefinition struct {
	Url      string      `json:"url"`
	Response string      `json:"response"`
	Status   int         `json:"status"`
	Method   string      `json:"method"`
	Headers  []APIHeader `json:"headers"`
}

func readFilesIntoDirectory(dir string) ([]fs.FileInfo, error) {
	os.Mkdir(dir, os.ModeDir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	//for _, file := range files {
	//	fmt.Println(file.Name())
	//}
	return files, err
}

func createFileIfNotExists(dir string, name string, content string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0777)
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile(dir+"/"+name, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(content)
	return nil
}

func readFile(dir string, name string) (string, error) {
	content, err := ioutil.ReadFile(dir + "/" + name)
	if err != nil {
		return "", err
	}
	return string(content), err
}

func createFolderIfNotExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func raw_connect(host string, ports []string) bool {
	for _, port := range ports {
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			//fmt.Println("Connecting error:", err)
			return false
		}
		if conn != nil {
			defer conn.Close()
			//fmt.Println("Opened", net.JoinHostPort(host, port))
			return true
		}
	}
	return false
}

func main() {
	// get port from os.Args --port=<port> or use default
	port := "8080"
	// Check args 1
	if len(os.Args) < 2 {
		// os.Exit(1)
	} else {
		port = os.Args[1]
	}
	// is port in use?
	// if yes, exit
	// if no, start server

	if port == "" {
		port = "8080"
	}

	ifPortInUse := raw_connect("localhost", []string{port})
	if ifPortInUse {
		fmt.Println("Port", port, "is in use")
		os.Exit(1)
	}

	path := "./services"
	createFolderIfNotExists(path)

	files, err := readFilesIntoDirectory(path)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	// print text with color
	fmt.Println("\033[1;32m[+]\033[0m", "Starting server on port "+port)
	fmt.Println("📂 Serving files from: " + path)
	fmt.Println("-----------------------------------------------------")
	fmt.Println("Allows for the following endpoints:")

	for _, file := range files {
		fileContent, err := readFile(path, file.Name())
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(fileContent)
		data := APIDefinition{}
		_ = json.Unmarshal([]byte(fileContent), &data)
		fmt.Println("\033[1;32m>\033[0m", data.Url, " - ", data.Method, " - ", data.Status)
		//fmt.Println(data.Response)

		r.HandleFunc(data.Url, func(w http.ResponseWriter, r *http.Request) {
			// assign header from data.headers
			for _, header := range data.Headers {
				w.Header().Set(header.Name, header.Value)
			}
			w.WriteHeader(data.Status)
			w.Write([]byte(data.Response))
		}).Methods(data.Method)
	}

	// create static routes
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"message\": \"Hello World\"}"))
	})

	http.ListenAndServe(":"+port, r)

	// createFileIfNotExists(path, "service.go", `{}`)
}
