package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"loggerator-go/reader"
)

const TEST_DATA_DIR = "testdata"
const DOT_LOG = ".log"
const DOT_GZ = ".gz"

func Test_200_logs_default_lines_parameter(t *testing.T) {
	filePath, err := createTestLogFile(TEST_DATA_DIR, DOT_LOG)
	require.NoError(t, err)
	defer func() {
		os.Remove(filePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))

	req, err := http.NewRequest("GET", "/logs", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func Test_200_logs_valid_lines_parameter(t *testing.T) {
	filePath, err := createTestLogFile(TEST_DATA_DIR, DOT_LOG)
	require.NoError(t, err)
	defer func() {
		os.Remove(filePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))

	req, err := http.NewRequest("GET", "/logs?lines=3", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func Test_200_logs_invalid_lines_parameter(t *testing.T) {
	filePath, err := createTestLogFile(TEST_DATA_DIR, DOT_LOG)
	require.NoError(t, err)
	defer func() {
		os.Remove(filePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))

	req, err := http.NewRequest("GET", "/logs?lines=invalid", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func Test_200_logs_all_lines_parameter(t *testing.T) {
	filePath, err := createTestLogFile(TEST_DATA_DIR, DOT_LOG)
	require.NoError(t, err)
	defer func() {
		os.Remove(filePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))
	req, err := http.NewRequest("GET", "/logs?lines=-1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func Test_200_logs_excluded_file_types(t *testing.T) {
	logFilePath, err := createTestLogFile(TEST_DATA_DIR, DOT_LOG)
	gzFilePath, err := createTestLogFile(TEST_DATA_DIR, DOT_GZ)
	require.NoError(t, err)
	defer func() {
		os.Remove(logFilePath)
		os.Remove(gzFilePath)
		os.Remove(TEST_DATA_DIR)
	}()

	handler := createLogsHandler(reader.NewReader(TEST_DATA_DIR))

	req, err := http.NewRequest("GET", "/logs?excludedFileTypes=gz", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
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

func createTestLogFile(dir string, extension string) (string, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	filePath := dir + "/test" + extension
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	for i := 0; i < 25; i++ {
		timestamp := time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339)
		_, err := file.WriteString(fmt.Sprintf("%s Log line %d\n", timestamp, i+1))
		if err != nil {
			return "", err
		}
	}

	return filePath, nil
}
