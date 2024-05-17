package check

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	buff       bytes.Buffer
	buffLogger = log.New(&buff, "", log.Lmsgprefix)
)

type BuffLogger struct{}

func (l BuffLogger) Print(str string) {
	buffLogger.Println(str)
}
func (l BuffLogger) Error(err error) {
	buffLogger.Println(err)
}
func (l BuffLogger) Success(str string) {
	buffLogger.Println(str)
}

func TestCheckInvalidFile(t *testing.T) {
	content := []byte(`Feature: test
  test

Scenario:            scenario1
  Given       whatever
  Then                  whatever
"""
hello world
"""

`)

	assert.NoError(t, os.RemoveAll("tmp/"))
	assert.NoError(t, os.MkdirAll("tmp/", 0o777))
	assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o777))

	command := NewCommand(BuffLogger{})
	command.SetArgs([]string{"tmp/file1.feature"})
	err := command.Execute()

	assert.Error(t, err)
	assert.Empty(t, buff.String())
	assert.EqualValues(t, `an error occurred with file "tmp/file1.feature" : file is not properly formatted`, err.Error())
	// Clean up
	_ = os.RemoveAll("tmp/")
	buff.Reset()
}

func TestCheckInvalidFolder(t *testing.T) {
	content := []byte(`Feature: test
  test

Scenario:            scenario1
  Given       whatever
  Then                  whatever
"""
hello world
"""

`)

	assert.NoError(t, os.RemoveAll("tmp/"))
	assert.NoError(t, os.MkdirAll("tmp/", 0o777))
	assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o777))

	command := NewCommand(BuffLogger{})
	command.SetArgs([]string{"tmp", "-i", "4"})
	err := command.Execute()

	assert.NoError(t, err)

	i := 0
	errs := strings.Split(buff.String(), "\n")
	expectedErrs := []string{
		`an error occurred with file "tmp/file1.feature" : file is not properly formatted`,
	}
	for _, expectedErr := range expectedErrs {
		for _, e := range errs {
			if expectedErr == e {
				i++
			}
		}
	}
	if l := len(expectedErrs); i != l {
		assert.Fail(t, fmt.Sprintf("Must fail with %v files when formatting folder", l))
	}
	// Clean up
	_ = os.RemoveAll("tmp/")
	buff.Reset()
}
