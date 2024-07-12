package check

import (
	"bytes"
	"os"
	"testing"

	"github.com/judimator/augurken/log"
	"github.com/stretchr/testify/assert"
)

func TestCheckInvalidFile(t *testing.T) {
	var buff bytes.Buffer
	logger := log.GetLogger()
	logger.SetOutput(&buff)

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
	assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o600))

	command := NewCommand()
	command.SetArgs([]string{"tmp/file1.feature"})
	err := command.Execute()

	assert.Error(t, err)
	assert.EqualValues(t, `an error occurred with file "tmp/file1.feature" : file is not properly formatted`+"\n", buff.String())
	// Clean up
	_ = os.RemoveAll("tmp/")
}

func TestCheckInvalidFolder(t *testing.T) {
	var buff bytes.Buffer
	logger := log.GetLogger()
	logger.SetOutput(&buff)

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
	assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o600))

	command := NewCommand()
	command.SetArgs([]string{"tmp", "-i", "4"})
	err := command.Execute()

	assert.Error(t, err)
	assert.EqualValues(t, `an error occurred with file "tmp/file1.feature" : file is not properly formatted`+"\n", buff.String())
	// Clean up
	_ = os.RemoveAll("tmp/")
}
