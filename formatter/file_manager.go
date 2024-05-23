package formatter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	mpath "path"
	"path/filepath"
	"sync"

	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

type FileManager struct {
	indent int
}

type ProcessFileError struct {
	Message string
	File    string
}

func (p ProcessFileError) Error() string {
	return fmt.Sprintf(`an error occurred with file "%s" : %s`, p.File, p.Message)
}

func NewFileManager(indent int) FileManager {
	return FileManager{
		indent,
	}
}

// FormatAndReplace Format and replace file or path. The function must return either []string or []error
func (f FileManager) FormatAndReplace(path string) []interface{} {
	return f.process(path, replaceFileWithContent)
}

// Check Test file or path. The function must return either []string or []error
func (f FileManager) Check(path string) []interface{} {
	return f.process(path, check)
}

func (f FileManager) Format(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}

	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(content)

	if err != nil {
		return []byte{}, err
	}

	if result.Charset != "UTF-8" {
		r, err := charset.NewReaderLabel(result.Charset, bytes.NewBuffer(content))
		if err != nil {
			return []byte{}, err
		}

		content, err = io.ReadAll(r)
		if err != nil {
			return []byte{}, err
		}
	}

	contentHelper := &ContentHelper{}
	contentHelper.DetectSettings(content)
	content = contentHelper.Prepare(content)

	token, err := parse(content)
	if err != nil {
		return []byte{}, err
	}

	return contentHelper.Restore(format(token, f.indent)), nil
}

// process Handle file or path depends on processFn value. The function must return either []string or []error
func (f FileManager) process(path string, processFn func(file string, content []byte) error) []interface{} {
	var result []interface{}
	fi, err := os.Stat(path)

	if err != nil {
		return append(result, err)
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		result = append(result, f.processPath(path, processFn)...)
	case mode.IsRegular():
		b, err := f.Format(path)
		if err != nil {
			return append(result, err)
		}

		if err := processFn(path, b); err != nil {
			return append(result, err)
		}

		result = append(result, fmt.Sprint("formatted: ", path))
	}

	return result
}

// processPath Handle path depends on processFn value. The function must return either []string or []error
func (f FileManager) processPath(path string, processFn func(file string, content []byte) error) []interface{} {
	var result []interface{}
	fc := make(chan string)
	wg := sync.WaitGroup{}

	files, err := findFeatureFiles(path)
	if err != nil {
		return append(result, err)
	}

	if len(files) == 0 {
		return result
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			for file := range fc {
				b, err := f.Format(file)

				if err != nil {
					result = append(result, ProcessFileError{Message: err.Error(), File: file})

					continue
				}

				if err := processFn(file, b); err != nil {
					result = append(result, err)

					continue
				}

				result = append(result, fmt.Sprint("formatted: ", file))
			}

			wg.Done()
		}()
	}

	for _, file := range files {
		fc <- file
	}

	close(fc)
	wg.Wait()

	return result
}

func replaceFileWithContent(file string, content []byte) error {
	f, err := os.Create(file)
	if err != nil {
		return ProcessFileError{Message: err.Error(), File: file}
	}

	_, err = f.Write(content)
	if err != nil {
		return ProcessFileError{Message: err.Error(), File: file}
	}

	return nil
}

func check(file string, content []byte) error {
	currentContent, err := os.ReadFile(file)
	if err != nil {
		return ProcessFileError{Message: err.Error(), File: file}
	}

	if !bytes.Equal(currentContent, content) {
		return ProcessFileError{Message: "file is not properly formatted", File: file}
	}

	return nil
}

func findFeatureFiles(rootPath string) ([]string, error) {
	var files []string

	if err := filepath.Walk(rootPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && mpath.Ext(p) == ".feature" {
			files = append(files, p)
		}

		return nil
	}); err != nil {
		return []string{}, err
	}

	return files, nil
}
