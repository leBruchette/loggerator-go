package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"loggerator-go/reader"
	"math"
	"net/http"
	"os"
	"strconv"
)

func main() {
	startServer(createFileReader())
}

func createFileReader() *reader.Reader {
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/var/log"
	}

	return reader.NewReader(logDir)
}

func startServer(fileReader *reader.Reader) {
	r := chi.NewRouter()
	r.Get("/logs", createLogsHandler(fileReader))

	logrus.Info("Server listening on port 8080...")
	http.ListenAndServe(":8080", r)
}

func createLogsHandler(fileReader *reader.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// number of lines to read from the log file
		linesToRead := parseLinesParameter(r)
		// comma-separated list of file extensions to exclude
		excludedFileTypes := r.URL.Query().Get("excludedFileTypes")
		// search text to filter log lines
		searchText := r.URL.Query().Get("search")

		fileContents, err := fileReader.GetLogFileContent(linesToRead, excludedFileTypes, searchText)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		transformOutput(w, err, fileContents)
	}
}

// parseLinesParameter parses the number of lines to read from the query parameter and returns the number of lines
// to read.
// If the query parameter is not provided or is invalid, the default number of lines to read is returned.
// If the query parameter is -1, all lines are read.
func parseLinesParameter(r *http.Request) int {
	defaultLineCount := 20
	var linesToRead int
	linesQueryParm := r.URL.Query().Get("lines")
	if len(linesQueryParm) == 0 {
		linesToRead = defaultLineCount
	} else if lines, err := strconv.Atoi(linesQueryParm); err != nil {
		logrus.Warningf("Invalid number of lines to read: '%s', defaulting to %d", linesQueryParm, defaultLineCount)
		linesToRead = defaultLineCount
	} else if lines == -1 {
		linesToRead = math.MaxInt
	} else {
		linesToRead = lines
	}
	return linesToRead
}

func transformOutput(w http.ResponseWriter, err error, fileContents []reader.FileContent) {
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(fileContents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
