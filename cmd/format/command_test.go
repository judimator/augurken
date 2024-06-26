package format

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatAndReplaceFile(t *testing.T) {
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
	command.SetArgs([]string{"tmp/file1.feature", "-i", "4"})
	err := command.Execute()

	assert.NoError(t, err)

	b, err := os.ReadFile("tmp/file1.feature")
	expected := `Feature: test
    test

    Scenario: scenario1
        Given whatever
        Then whatever
            """
            hello world
            """

`

	assert.NoError(t, err)
	assert.EqualValues(t, expected, string(b))

	// Clean up
	_ = os.RemoveAll("tmp/")
}

func TestFormatAndReplaceFolder(t *testing.T) {
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

	assert.NoError(t, err)

	b, err := os.ReadFile("tmp/file1.feature")
	expected := `Feature: test
    test

    Scenario: scenario1
        Given whatever
        Then whatever
            """
            hello world
            """

`

	assert.NoError(t, err)
	assert.EqualValues(t, expected, string(b))

	// Clean up
	_ = os.RemoveAll("tmp/")
}
