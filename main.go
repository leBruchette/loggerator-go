package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"loggerator-go/reader"
	"net/http"
	"os"
)

var (
	logDir     string
	fileReader *reader.Reader
)

func main() {
	// initialize app context
	logDir = os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/var/log"
	}

	fileReader = reader.NewReader(logDir)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/logs", getLogsContent)

	logrus.Info("Server listening on port 8080...")
	http.ListenAndServe(":8080", r)
}

func getLogsContent(w http.ResponseWriter, r *http.Request) {
	//comma-separated list of file extensions to exclude
	excludedFileTypes := r.URL.Query().Get("excludedFileTypes")

	fileContents, err := fileReader.GetLogFileContent(5, excludedFileTypes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	funcName(w, err, fileContents)
}

func funcName(w http.ResponseWriter, err error, fileContents map[string]reader.FileContent) {
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(fileContents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
