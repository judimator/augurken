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
	logger Log
}

type ProcessFileError struct {
	Message string
	File    string
}

func (p ProcessFileError) Error() string {
	return fmt.Sprintf(`an error occurred with file "%s" : %s`, p.File, p.Message)
}

func NewFileManager(indent int, logger Log) FileManager {
	return FileManager{
		indent,
		logger,
	}
}

func (f FileManager) FormatAndReplace(path string) error {
	return f.process(path)
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

	content, err = format(token, f.indent)
	if err != nil {
		return []byte{}, err
	}

	return contentHelper.Restore(content), nil
}

func (f FileManager) process(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		if err := f.processPath(path); err != nil {
			return err
		}
	case mode.IsRegular():
		b, err := f.Format(path)
		if err != nil {
			return err
		}

		if err := replaceFileWithContent(path, b); err != nil {
			return err
		}

		f.logger.Success(fmt.Sprint("+", path))
	}
	return nil
}

func (f FileManager) processPath(path string) error {
	fc := make(chan string)
	wg := sync.WaitGroup{}

	files, err := findFeatureFiles(path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for file := range fc {
				b, err := f.Format(file)

				if err != nil {
					f.logger.Error(ProcessFileError{Message: err.Error(), File: file})
					continue
				}

				if err := replaceFileWithContent(file, b); err != nil {
					f.logger.Error(ProcessFileError{Message: err.Error(), File: file})
					continue
				}

				f.logger.Success(fmt.Sprint("formatted: ", file))
			}
			wg.Done()
		}()
	}
	for _, file := range files {
		fc <- file
	}

	close(fc)
	wg.Wait()
	return nil
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
