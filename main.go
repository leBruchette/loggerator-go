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
	"strings"
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

// the main, aggregation server that queries the managed server is ec2-18-117-92-75.us-east-2.compute.amazonaws.com
// e.g curl --location 'http://ec2-18-117-92-75.us-east-2.compute.amazonaws.com:8080/managed/logs'
func createManagedLogsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers := []string{
			"ec2-3-147-61-154.us-east-2.compute.amazonaws.com",
			"ec2-18-224-140-50.us-east-2.compute.amazonaws.com",
			"ec2-3-16-23-61.us-east-2.compute.amazonaws.com",
			"ec2-3-144-37-137.us-east-2.compute.amazonaws.com",
			"ec2-3-17-11-20.us-east-2.compute.amazonaws.com",
		}

		// channel for receiving results from the servers
		resultsChan := make(chan *reader.ServerContent, len(servers))
		var wg sync.WaitGroup

		// remove `/managed` and pass through as `/logs` with any query parameters present
		endpoint := strings.Replace(r.URL.Path, "/managed", "", 1)
		if rawQuery := r.URL.RawQuery; rawQuery != "" {
			endpoint += "?" + rawQuery
		}

		for _, server := range servers {
			wg.Add(1)
			go func(server string) {
				defer wg.Done()
				url := "http://" + server + ":8080" + endpoint
				resp, err := http.Get(url)
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
				resultsChan <- &reader.ServerContent{HostName: server, Files: fileContents}
			}(server)
		}

		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// collect results from the servers
		var results []*reader.ServerContent
		for result := range resultsChan {
			results = append(results, result)
		}

		w.Header().Set("Content-Type", "application/json")
		if len(results) == 0 {
			http.ResponseWriter.WriteHeader(w, http.StatusNoContent)
		}
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
