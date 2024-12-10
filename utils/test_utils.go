//go:build !testcoverage

package utils

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"time"
)

func CreateTestLogFile(dir string, extension string) string {
	os.MkdirAll(dir, 0755)

	filePath := dir + "/file" + extension
	file, _ := os.Create(filePath)
	defer file.Close()

	for i := 0; i < 25; i++ {
		timestamp := time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339)
		file.WriteString(fmt.Sprintf("%s Log line %d\n", timestamp, i+1))
	}

	return filePath
}

func CreateTestDataDir() string {
	testDataDir := "testdata"
	os.Mkdir(testDataDir, 0755)

	return testDataDir
}

func CreateEmptyDirectory() string {
	emptyDirPath := "testdata/empty"
	os.Mkdir(emptyDirPath, 0755)

	return emptyDirPath
}

func CreateEmptyFile() string {
	emptyFilePath := "testdata/empty.log"
	os.WriteFile(emptyFilePath, []byte(""), 0644)

	return emptyFilePath
}

func CreateRestrictedFile() string {
	// Create a non-readable file
	restrictedFilePath := "testdata/restricted.log"
	os.WriteFile(restrictedFilePath, []byte("nobody will ever know..."), 0644)
	os.Chmod(restrictedFilePath, 0000)

	return restrictedFilePath
}

func CreateCompressedFileWithExtension(src string, extension string) string {
	dest := src + extension
	file, _ := os.Open(src)
	defer file.Close()

	outFile, _ := os.Create(dest)
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()
	io.Copy(gw, file)
	return dest
}
