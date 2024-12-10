package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"loggerator-go/reader"
	"loggerator-go/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const TEST_DATA_DIR = "testdata"
const DOT_LOG = ".log"
const DOT_GZ = ".gz"

func Test_200_logs_default_lines_parameter(t *testing.T) {
	filePath := utils.CreateTestLogFile(TEST_DATA_DIR, DOT_LOG)
	defer func() {
		os.Remove(filePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))

	req, err := http.NewRequest("GET", "/logs", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var fileContents []reader.FileContent
	err = json.Unmarshal(rr.Body.Bytes(), &fileContents)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", strings.ToLower(rr.Header().Get("Content-Type")))
	assert.Equal(t, 1, len(fileContents))
	assert.Equal(t, "testdata/file.log", fileContents[0].Name)
	assert.Equal(t, 20, len(fileContents[0].Content))
}

func Test_200_logs_valid_lines_parameter(t *testing.T) {
	filePath := utils.CreateTestLogFile(TEST_DATA_DIR, DOT_LOG)
	defer func() {
		os.Remove(filePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))

	req, err := http.NewRequest("GET", "/logs?lines=3", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var fileContents []reader.FileContent
	err = json.Unmarshal(rr.Body.Bytes(), &fileContents)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", strings.ToLower(rr.Header().Get("Content-Type")))
	assert.Equal(t, 1, len(fileContents))
	assert.Equal(t, "testdata/file.log", fileContents[0].Name)
	assert.Equal(t, 3, len(fileContents[0].Content))
}

func Test_200_logs_invalid_lines_parameter(t *testing.T) {
	filePath := utils.CreateTestLogFile(TEST_DATA_DIR, DOT_LOG)
	defer func() {
		os.Remove(filePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))

	req, err := http.NewRequest("GET", "/logs?lines=invalid", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var fileContents []reader.FileContent
	err = json.Unmarshal(rr.Body.Bytes(), &fileContents)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", strings.ToLower(rr.Header().Get("Content-Type")))
	assert.Equal(t, 1, len(fileContents))
	assert.Equal(t, "testdata/file.log", fileContents[0].Name)
	assert.Equal(t, 20, len(fileContents[0].Content))
}

func Test_200_logs_all_lines_parameter(t *testing.T) {
	filePath := utils.CreateTestLogFile(TEST_DATA_DIR, DOT_LOG)
	defer func() {
		os.Remove(filePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))
	req, err := http.NewRequest("GET", "/logs?lines=-1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var fileContents []reader.FileContent
	err = json.Unmarshal(rr.Body.Bytes(), &fileContents)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", strings.ToLower(rr.Header().Get("Content-Type")))
	assert.Equal(t, 1, len(fileContents))
	assert.Equal(t, "testdata/file.log", fileContents[0].Name)
	assert.Equal(t, 25, len(fileContents[0].Content))
}

func Test_200_logs_excluded_file_types(t *testing.T) {
	logFilePath := utils.CreateTestLogFile(TEST_DATA_DIR, DOT_LOG)
	gzFilePath := utils.CreateCompressedFileWithExtension(logFilePath, DOT_GZ)
	defer func() {
		os.Remove(logFilePath)
		os.Remove(gzFilePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))

	req, err := http.NewRequest("GET", "/logs?excludedFileTypes=.log", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var fileContents []reader.FileContent
	err = json.Unmarshal(rr.Body.Bytes(), &fileContents)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", strings.ToLower(rr.Header().Get("Content-Type")))
	assert.Equal(t, 1, len(fileContents))
	assert.Equal(t, "testdata/file.log.gz", fileContents[0].Name)
	assert.Equal(t, 1, len(fileContents[0].Content))
}

func Test_createFileReader_DefaultLogDir(t *testing.T) {
	os.Setenv("LOG_DIR", "")
	defer os.Unsetenv("LOG_DIR")

	fileReader := createFileReader()
	assert.Equal(t, "/var/log", fileReader.Dir)
}

func Test_createFileReader_CustomLogDir(t *testing.T) {
	os.Setenv("LOG_DIR", "/custom/log/dir")
	defer os.Unsetenv("LOG_DIR")

	fileReader := createFileReader()
	assert.Equal(t, "/custom/log/dir", fileReader.Dir)
}
