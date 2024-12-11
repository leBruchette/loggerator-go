package reader

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	UnreadableFileMessage = "File is not human-readable"
	ChunkSize             = 1024 // Size of each chunk to read
)

type FileContent struct {
	Name     string    `json:"fileName"`
	Size     int64     `json:"fileSizeBytes"`
	Utf8     bool      `json:"isUtf8"`
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

func (r *Reader) GetLogFileContent(lineCount int, excluded string, searchText string) ([]FileContent, error) {
	// reads all files in the directory
	fileInfos, err := getFilesInDirectory(r)
	if err != nil {
		return nil, err
	}

	// create a channel to receive fileInfo contents
	fileContents := make(chan FileContent)
	var wg sync.WaitGroup

	// filter out directories and fileInfo extensions to exclude
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && !isExcluded(filepath.Ext(fileInfo.Name()), excluded) {
			// increment the number of goroutines to wait for
			wg.Add(1)

			// perform fileInfo manipulation within a goroutine
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
				readableContent, isUtf8, err := readLinesInReverse(f, lineCount, searchText)
				if err != nil {
					fileContents <- FileContent{Name: r.Dir + "/" + file.Name(), Err: err}
					return
				}
				// readableContent will be empty if a search term is not found, so don't return the fileInfo in the list
				if len(readableContent) > 0 {
					fileContents <- FileContent{
						Name:     r.Dir + "/" + file.Name(),
						Size:     file.Size(),
						Utf8:     isUtf8,
						Modified: file.ModTime().UTC(),
						Content:  readableContent,
						Err:      nil,
					}
				}
			}(fileInfo)
		}
	}

	// close the channel once all goroutines are done
	go func() {
		wg.Wait()
		close(fileContents)
	}()

	var sortedContents []FileContent
	for content := range fileContents {
		// don't return non-readable results if a search term is provided
		// just to narrow-down any compressed files that shouldn't be returned if a user is looking for a specific term
		// perhaps compressed files should be ignored altogether?
		if !content.Utf8 && len(searchText) > 0 {
			continue
		}
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

func readLinesInReverse(file *os.File, lineCount int, searchText string) ([]string, bool, error) {
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return nil, false, err
	}

	fileSize := fileStat.Size()
	buffer := make([]byte, ChunkSize)
	var lines []string
	var leftOver []byte
	for offset := fileSize; offset > 0 && len(lines) < lineCount; offset -= ChunkSize {
		// Adjust chunk size for the first chunk if smaller than ChunkSize
		if offset < ChunkSize {
			buffer = make([]byte, offset)
		}

		// Move the offset back to read the next chunk
		_, err := file.Seek(offset-int64(len(buffer)), io.SeekStart)
		if err != nil {
			return nil, false, err
		}

		_, err = file.Read(buffer)
		if err != nil {
			return nil, false, err
		}

		if !isHumanReadable(buffer) {
			return []string{UnreadableFileMessage}, false, nil
		}

		// Combine leftover from previous chunk
		combined := append(buffer, leftOver...)
		linesInChunk := bytes.Split(combined, []byte("\n"))

		// Keep the first part for next chunk
		if offset != ChunkSize {
			leftOver = linesInChunk[0]
		} else {
			leftOver = nil
		}

		for i := len(linesInChunk) - 1; i > 0 && len(lines) < lineCount; i-- {
			sanitizedLine := strings.ReplaceAll(string(linesInChunk[i]), "\u0000", "")
			if nonBlankLineOrContainsSearchText(sanitizedLine, searchText) {
				lines = append(lines, sanitizedLine)
			}
		}
	}

	// Add the last left over line if we still need more lines, but make sure it's not a blank line or doesn't contain the search text if provided
	if len(leftOver) > 0 && len(lines) < lineCount {
		sanitizedLine := strings.ReplaceAll(string(leftOver), "\u0000", "")
		if nonBlankLineOrContainsSearchText(sanitizedLine, searchText) {
			lines = append(lines, string(leftOver))
		}
	}

	return lines, true, nil
}

func nonBlankLineOrContainsSearchText(currLine string, searchText string) bool {
	return len(currLine) > 0 && (len(searchText) == 0 || strings.Contains(currLine, searchText))
}
