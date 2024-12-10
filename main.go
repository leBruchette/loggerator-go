package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"loggerator-go/reader"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all available CPUs
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
	r.Get("/status", createStatusHandler())
	r.Get("/logs", createLogsHandler(fileReader))
	r.Get("/managed/logs", createManagedLogsHandler())

	logrus.Info("Server listening on port 8080...")
	http.ListenAndServe(":8080", r)
}

func createStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Info("Received status request")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
	}
}

func createManagedLogsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers := []string{"server1.example.com", "server2.example.com"} // List of other servers
		results := make(map[string][]reader.FileContent)
		var mu sync.Mutex
		var wg sync.WaitGroup

		for _, server := range servers {
			wg.Add(1)
			go func(server string) {
				defer wg.Done()
				resp, err := http.Get("http://" + server + ":8080/" + r.URL.Path)
				if err != nil {
					logrus.Errorf("Error calling server %s: %v", server, err)
					return
				}
				defer resp.Body.Close()

				var fileContents []reader.FileContent
				if err := json.NewDecoder(resp.Body).Decode(&fileContents); err != nil {
					logrus.Errorf("Error decoding response from server %s: %v", server, err)
					return
				}

				mu.Lock()
				results[server] = fileContents
				mu.Unlock()
			}(server)
		}

		wg.Wait()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
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
