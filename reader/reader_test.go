package reader

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const EXCLUDED_FILE_TYPES = ""

func Test_GetLogFileContent_ReturnsCorrectContent(t *testing.T) {
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
	assert.Equal(t, 0, len(contents["testdata/empty.log"].Content))
	// gzipped file, or any file that is not in a human-readable format
	assert.Equal(t, 1, len(contents["testdata/file.log.gz"].Content))
	assert.Equal(t, "File is not human-readable", contents["testdata/file.log.gz"].Content[0])
}

func Test_GetLogFileContent_ExcludesSpecifiedFileExtensions(t *testing.T) {
	r := NewReader("testdata")
	contents, err := r.GetLogFileContent(5, ".gz")
	require.NoError(t, err)

	assert.Equal(t, 2, len(contents))
	assert.Contains(t, contents, "testdata/file.log")
	assert.Contains(t, contents, "testdata/empty.log")
	assert.NotContains(t, contents, "testdata/file.log.gz")

}

func Test_GetLogFileContent_HandlesEmptyDirectory(t *testing.T) {
	r := NewReader("testdata/empty")
	contents, err := r.GetLogFileContent(5, EXCLUDED_FILE_TYPES)
	require.NoError(t, err)

	assert.Empty(t, contents)
}

func Test_GetLogFileContent_HandlesNonReadableFiles(t *testing.T) {
	r := NewReader("testdata/restricted")
	contents, err := r.GetLogFileContent(5, EXCLUDED_FILE_TYPES)
	require.NoError(t, err)

	assert.Equal(t, 2, len(contents))
	assert.Equal(t, 5, len(contents["testdata/restricted/file.log"].Content))
	assert.Empty(t, contents["testdata/restricted/file.log"].Err)

	// restricted file cannot be read/searched
	assert.Equal(t, 0, len(contents["testdata/restricted/restricted.log"].Content))
	assert.Contains(t, contents["testdata/restricted/restricted.log"].Err.Error(), "permission denied")
}

func Test_GetLogFileContent_ErrorsOnNonExistentDirectory(t *testing.T) {
	r := NewReader("imnothere")
	_, err := r.GetLogFileContent(5, EXCLUDED_FILE_TYPES)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}
