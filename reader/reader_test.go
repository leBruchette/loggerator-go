package reader

import (
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

const EXCLUDED_FILE_TYPES = ""

func Test_GetLogFileContent_ReturnsCorrectContent(t *testing.T) {
	emptyFilePath, err := createEmptyFile(t)
	require.NoError(t, err)
	gzippedFilePath, err := createCompressedFileWithExtension("testdata/file.log", ".gz")
	require.NoError(t, err)
	defer func() {
		os.Remove(emptyFilePath)
		os.Remove(gzippedFilePath)
	}()

	r := NewReader("testdata")
	contents, err := r.GetLogFileContent(5, EXCLUDED_FILE_TYPES)
	require.NoError(t, err)

	assert.Equal(t, 3, len(contents))
	// normal, readable file
	assert.Equal(t, 5, len(contents["testdata/file.log"].Content))
	assert.Equal(t, "2024-12-16 eighth", contents["testdata/file.log"].Content[0])
	assert.Equal(t, "2024-12-15 seventh", contents["testdata/file.log"].Content[1])
	assert.Equal(t, "2024-12-14 sixth", contents["testdata/file.log"].Content[2])
	assert.Equal(t, "2024-12-13 fifth", contents["testdata/file.log"].Content[3])
	assert.Equal(t, "2024-12-12 fourth", contents["testdata/file.log"].Content[4])
	// empty log file
	assert.Equal(t, 0, len(contents[emptyFilePath].Content))
	// gzipped file, or any file that is not in a human-readable format
	assert.Equal(t, 1, len(contents[gzippedFilePath].Content))
	assert.Equal(t, "File is not human-readable", contents[gzippedFilePath].Content[0])
}

func Test_GetLogFileContent_ExcludesSpecifiedFileExtensions(t *testing.T) {
	gzippedFilePath, err := createCompressedFileWithExtension("testdata/file.log", ".gz")
	require.NoError(t, err)
	derpedFilePath, err := createCompressedFileWithExtension("testdata/file.log", ".derp")
	require.NoError(t, err)
	defer func() {
		os.Remove(gzippedFilePath)
		os.Remove(derpedFilePath)
	}()

	r := NewReader("testdata")
	contents, err := r.GetLogFileContent(5, ".gz,.derp")
	require.NoError(t, err)

	assert.Equal(t, 1, len(contents))
	assert.Contains(t, contents, "testdata/file.log")
	assert.NotContains(t, contents, "testdata/file.log.gz")
	assert.NotContains(t, contents, "testdata/file.log.derp")

}

func Test_GetLogFileContent_HandlesEmptyDirectory(t *testing.T) {
	emptyDirPath, err := createEmptyDirectory(t)
	require.NoError(t, err)
	defer func() {
		os.Remove(emptyDirPath)
	}()

	r := NewReader(emptyDirPath)
	contents, err := r.GetLogFileContent(5, EXCLUDED_FILE_TYPES)
	require.NoError(t, err)

	assert.Empty(t, contents)
}

func Test_GetLogFileContent_HandlesNonReadableFiles(t *testing.T) {
	restrictedFilePath, err := createRestrictedFile(t)
	require.NoError(t, err)
	defer func() {
		os.Chmod(restrictedFilePath, 0644)
		os.Remove(restrictedFilePath)
	}()

	r := NewReader("testdata")
	contents, err := r.GetLogFileContent(5, EXCLUDED_FILE_TYPES)
	require.NoError(t, err)

	assert.Equal(t, 2, len(contents))
	assert.Equal(t, 5, len(contents["testdata/file.log"].Content))
	assert.Empty(t, contents["testdata/file.log"].Err)
	// restricted file cannot be read/searched
	assert.Equal(t, 0, len(contents[restrictedFilePath].Content))
	assert.Contains(t, contents[restrictedFilePath].Err.Error(), "permission denied")
}

func Test_GetLogFileContent_ErrorsOnNonExistentDirectory(t *testing.T) {
	r := NewReader("imnothere")
	_, err := r.GetLogFileContent(5, EXCLUDED_FILE_TYPES)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func createEmptyFile(t *testing.T) (string, error) {
	emptyFilePath := "testdata/empty.log"
	err := os.WriteFile(emptyFilePath, []byte(""), 0644)
	require.NoError(t, err)
	return emptyFilePath, err
}

func createRestrictedFile(t *testing.T) (string, error) {
	// Create a non-readable file
	restrictedFilePath := "testdata/restricted.log"
	err := os.WriteFile(restrictedFilePath, []byte("nobody will ever know..."), 0644)
	require.NoError(t, err)
	err = os.Chmod(restrictedFilePath, 0000)
	require.NoError(t, err)

	return restrictedFilePath, err
}

func createEmptyDirectory(t *testing.T) (string, error) {
	emptyDirPath := "testdata/empty"
	err := os.Mkdir(emptyDirPath, 0755)
	require.NoError(t, err)
	return emptyDirPath, err
}

func createCompressedFileWithExtension(src string, extension string) (string, error) {
	dest := src + extension
	file, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer file.Close()
	outFile, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	gw := gzip.NewWriter(outFile)
	defer gw.Close()
	if _, err := io.Copy(gw, file); err != nil {
		return "", err
	}
	return dest, nil
}
