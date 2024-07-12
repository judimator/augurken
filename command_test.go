package main

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Main(t *testing.T) {
	os.Args = []string{"augurken", "--help"}
	exitFn = func(code int) { assert.Equal(t, 0, code) }

	r, w, _ := os.Pipe()
	os.Stdout = w

	main()
	_ = w.Close()
	buf := new(bytes.Buffer)

	_ = r.SetReadDeadline(time.Now().Add(time.Second))
	_, _ = io.Copy(buf, r)

	assert.Contains(t, buf.String(), "Usage:")
	assert.Contains(t, buf.String(), "Available Commands:")
	assert.Contains(t, buf.String(), "Flags:")
}

func Test_MainWithoutCommands(t *testing.T) {
	os.Args = []string{"augurken"}
	exitFn = func(code int) { assert.Equal(t, 0, code) }

	r, w, _ := os.Pipe()
	os.Stdout = w

	main()
	buf := new(bytes.Buffer)
	_ = r.SetReadDeadline(time.Now().Add(time.Second))
	_, _ = io.Copy(buf, r)

	assert.Contains(t, buf.String(), "Usage:")
	assert.Contains(t, buf.String(), "Available Commands:")
	assert.Contains(t, buf.String(), "Flags:")
}

func Test_MainUnknownSubcommand(t *testing.T) {
	os.Args = []string{"", "foobar"}
	exitFn = func(code int) { assert.Equal(t, 1, code) }

	r, w, _ := os.Pipe()
	os.Stderr = w

	main()
	_ = w.Close()
	buf := new(bytes.Buffer)

	_ = r.SetReadDeadline(time.Now().Add(time.Second))
	_, _ = io.Copy(buf, r)

	assert.Contains(t, buf.String(), "unknown command")
	assert.Contains(t, buf.String(), "foobar")
}

func Test_MainCheckFolderError(t *testing.T) {
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

	// Set up the command
	os.Args = []string{"augurken", "check", "tmp"}
	exitFn = func(code int) { assert.Equal(t, 1, code) }

	r, w, _ := os.Pipe()
	os.Stderr = w

	main()
	_ = w.Close()
	buf := new(bytes.Buffer)

	_ = r.SetReadDeadline(time.Now().Add(time.Second))
	_, _ = io.Copy(buf, r)

	assert.Contains(t, buf.String(), "Error: error occurred while formatting file/folder")

	// Clean up
	_ = os.RemoveAll("tmp/")
}
