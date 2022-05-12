package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type APIDefinition struct {
	Url      string `json:"url"`
	Response string `json:"response"`
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

func main() {
	path := "./services"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	files, err := readFilesIntoDirectory(path)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	// print text with color
	fmt.Println("\033[1;32m[+]\033[0m", "Starting server on port 8080")
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
		fmt.Println("\033[1;32m>\033[0m", data.Url)
		//fmt.Println(data.Response)

		r.HandleFunc(data.Url, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(data.Response))
		})
	}

	// create static routes
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"message\": \"Hello World\"}"))
	})

	http.ListenAndServe(":8080", r)
	// createFileIfNotExists(path, "service.go", `{}`)
}
