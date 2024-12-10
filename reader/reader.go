package reader

import (
	"bufio"
	"bytes"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type FileContent struct {
	Name     string    `json:"fileName"`
	Size     int64     `json:"fileSizeBytes"`
	Modified time.Time `json:"lastModified"`
	Content  []string  `json:"content,omitempty"`
	Err      error     `json:"error,omitempty"`
}

type Reader struct {
	Dir string
}

func NewReader(dir string) *Reader {
	return &Reader{
		Dir: dir,
	}
}

func isExcluded(fileType string, excludedFileTypes string) bool {
	if len(excludedFileTypes) == 0 {
		return false
	}

	for _, e := range strings.Split(excludedFileTypes, ",") {
		if fileType == e {
			return true
		}
	}
	return false
}

func (r *Reader) GetLogFileContent(lineCount int, excluded string) ([]FileContent, error) {

	// reads all files in the directory
	fileInfos, err := getFilesInDirectory(r)
	if err != nil {
		return nil, err
	}

	// create a channel to receive file contents
	fileContents := make(chan FileContent)
	var wg sync.WaitGroup

	// filter out directories and file extensions to exclude
	for _, file := range fileInfos {
		if !file.IsDir() && !isExcluded(filepath.Ext(file.Name()), excluded) {
			// increment the number of goroutines to wait for
			wg.Add(1)

			// perform file manipulation within a goroutine
			go func(file os.FileInfo) {
				defer wg.Done()
				path := filepath.Join(r.Dir, file.Name())
				f, err := os.Open(path)
				if err != nil {

					fileContents <- FileContent{Name: r.Dir + "/" + file.Name(), Err: err}
					return
				}
				defer f.Close()

				var readableContent []string
				readableContent = readLinesReverse(f, lineCount)
				fileContents <- FileContent{
					Name:     r.Dir + "/" + file.Name(),
					Size:     file.Size(),
					Modified: file.ModTime().UTC(),
					Content:  readableContent,
					Err:      nil,
				}
			}(file)
		}
	}

	// close the channel once all goroutines are done
	go func() {
		wg.Wait()
		close(fileContents)
	}()

	var sortedContents []FileContent
	for content := range fileContents {
		sortedContents = append(sortedContents, content)
	}

	// Sort the contents by Modified time
	sort.Slice(sortedContents, func(i, j int) bool {
		return sortedContents[i].Modified.After(sortedContents[j].Modified)
	})

	return sortedContents, nil
}

func getFilesInDirectory(r *Reader) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(r.Dir)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	var fileInfos []os.FileInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				// don't fail everything because of one bad file...
				logrus.Error("Error reading file entry, skipping: ", err)
				continue
			}
			fileInfos = append(fileInfos, info)
		}
	}
	return fileInfos, nil
}

func isHumanReadable(content []byte) bool {
	for i := 0; i < len(content); {
		r, size := utf8.DecodeRune(content[i:])
		if r == utf8.RuneError && size == 1 {
			return false
		}
		i += size
	}
	return true
}

func readLinesReverse(file *os.File, lineCount int) []string {
	const bufferSize = 4096
	stat, err := file.Stat()
	if err != nil {
		log.Fatalf("failed to get file info: %v", err)
	}
	fileSize := stat.Size()
	buffer := make([]byte, bufferSize)

	initialReadSize := bufferSize
	if fileSize < int64(bufferSize) {
		initialReadSize = int(fileSize)
	}

	// Fail fast if the file can't be read
	_, err = file.ReadAt(buffer[:initialReadSize], 0)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}
	// Read the first part of the file to check if it's human-readable
	// if not, return a slice with a relevant message
	if !isHumanReadable(buffer[:initialReadSize]) {
		return []string{"File is not human-readable"}
	}

	var lines []string
	for offset := fileSize; offset > 0; {
		toRead := int64(bufferSize)
		if offset < bufferSize {
			toRead = offset
		}
		offset -= toRead
		file.Seek(offset, io.SeekStart)
		n, err := file.Read(buffer[:toRead])
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		bufferText := buffer[:n]
		// Handling partial lines at the edges
		if len(lines) > 0 {
			bufferText = append(bufferText, []byte(lines[0])...)
			lines = lines[1:]
		}

		scanner := bufio.NewScanner(bytes.NewReader(bufferText))
		// Create a 1MB slice to handle large lines
		const maxCapacity = 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		var tempLines []string
		for scanner.Scan() {
			//filter out empty lines
			currLine := scanner.Text()
			if len(currLine) > 0 {
				tempLines = append(tempLines, currLine)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("error scanning file %v: %v", file.Name(), err)
		}
		// Reverse order of lines read in this iteration
		for i := len(tempLines) - 1; i >= 0; i-- {
			lines = append(lines, tempLines[i])
			if len(lines) >= lineCount {
				break
			}
		}
	}

	return lines
}
