package reader

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"loggerator-go/utils"
	"os"
	"testing"
)

const (
	EmptyExclusionList = ""
	LogFileDir         = "testdata"
)

var (
	testDataDir        string
	emptyFilePath      string
	emptyDirPath       string
	restrictedFilePath string
	gzippedFilePath    string
	testLogFilePath    string
)

func setup() {
	fmt.Println("Setup: Running before all tests")
	// creating these dirs/file in order since we're sorting by modified date
	testDataDir = utils.CreateTestDataDir()
	emptyDirPath = utils.CreateEmptyDirectory()
	restrictedFilePath = utils.CreateRestrictedFile() // oldest modified file
	emptyFilePath = utils.CreateEmptyFile()
	gzippedFilePath = utils.CreateCompressedFileWithExtension("testdata/file.log", ".gz")
	testLogFilePath = utils.CreateTestLogFile(LogFileDir, ".log") // latest modified file
}

func teardown() {
	fmt.Println("Teardown: Running after all tests")
	defer func() {
		os.Remove(testLogFilePath)
		os.Remove(gzippedFilePath)
		os.Remove(emptyFilePath)
		os.Chmod(restrictedFilePath, 0644)
		os.Remove(restrictedFilePath)
		os.Remove(emptyDirPath)
		os.Remove(testDataDir)
	}()
}

// TestMain is the entry point for the tests in this package
func TestMain(m *testing.M) {
	setup()
	exitVal := m.Run()
	teardown()

	os.Exit(exitVal)
}

func Test_GetLogFileContent_ReturnsCorrectContent(t *testing.T) {
	r := NewReader(LogFileDir)
	contents, err := r.GetLogFileContent(5, EmptyExclusionList)
	require.NoError(t, err)

	assert.Equal(t, 4, len(contents))
	assert.True(t, contents[0].Modified.After(contents[1].Modified))
	assert.True(t, contents[1].Modified.After(contents[2].Modified))
	assert.True(t, contents[2].Modified.After(contents[3].Modified))
	// testLogFile
	assert.Equal(t, "testdata/file.log", contents[0].Name)
	assert.Equal(t, 5, len(contents[0].Content))
	assert.Contains(t, contents[0].Content[0], "Log line 25")
	assert.Contains(t, contents[0].Content[1], "Log line 24")
	assert.Contains(t, contents[0].Content[2], "Log line 23")
	assert.Contains(t, contents[0].Content[3], "Log line 22")
	assert.Contains(t, contents[0].Content[4], "Log line 21")
	// gzipped file,
	assert.Equal(t, "testdata/file.log.gz", contents[1].Name)
	assert.Equal(t, 1, len(contents[1].Content))
	assert.Equal(t, "File is not human-readable", contents[1].Content[0])
	// empty file,
	assert.Equal(t, "testdata/empty.log", contents[2].Name)
	assert.Equal(t, 0, len(contents[2].Content))
	// restrictedLogFile
	assert.Equal(t, "testdata/restricted.log", contents[3].Name)
	assert.Equal(t, 0, len(contents[3].Content))
}

func Test_GetLogFileContent_ExcludesSpecifiedFileExtensions(t *testing.T) {
	r := NewReader(LogFileDir)
	contents, err := r.GetLogFileContent(-1, ".gz")
	require.NoError(t, err)

	assert.Equal(t, 3, len(contents))
	assert.Equal(t, contents[0].Name, "testdata/file.log")
	assert.Equal(t, contents[1].Name, "testdata/empty.log")
	assert.Equal(t, contents[2].Name, "testdata/restricted.log")
}

func Test_GetLogFileContent_HandlesEmptyDirectory(t *testing.T) {
	r := NewReader(emptyDirPath)
	contents, err := r.GetLogFileContent(5, EmptyExclusionList)
	require.NoError(t, err)

	assert.Empty(t, contents)
}

func Test_GetLogFileContent_HandlesNonReadableFiles(t *testing.T) {
	r := NewReader(LogFileDir)
	contents, err := r.GetLogFileContent(5, ".log")
	require.NoError(t, err)

	assert.Equal(t, 1, len(contents))
	// gzipped file is not readable
	assert.Equal(t, "testdata/file.log.gz", contents[0].Name)
	assert.Equal(t, 1, len(contents[0].Content))
	assert.Equal(t, "File is not human-readable", contents[0].Content[0])
}

func Test_GetLogFileContent_ErrorsOnNonExistentDirectory(t *testing.T) {
	r := NewReader("imnothere")
	_, err := r.GetLogFileContent(5, EmptyExclusionList)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}
