package formatter

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileManagerFormat(t *testing.T) { //nolint:tparallel
	type scenario struct {
		filename string
		test     func([]byte, error)
	}

	scenarios := []scenario{
		{
			"features/comment-after-background.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/comment-after-background.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/comment-after-scenario.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/comment-after-scenario.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/comment-in-a-midst-of-row.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/comment-in-a-midst-of-row.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/docstring-empty.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/docstring-empty.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/escape-new-line.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/escape-new-line.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/escape-pipe.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/escape-pipe.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/file1.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/file1.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/utf8-with-bom.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/utf8-with-bom.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/file1-with-cr.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/file1-with-cr.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/file1-with-crlf.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/file1-with-crlf.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/iso-8859-1-encoding.input.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/iso-8859-1-encoding.expected.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/several-scenario-following.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)

				b, e := os.ReadFile("features/several-scenario-following.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"features/",
			func(_ []byte, err error) {
				assert.EqualError(t, err, "read features/: is a directory")
			},
		},
		{
			"features/invalid.feature",
			func(_ []byte, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.filename, func(t *testing.T) {
			t.Parallel()
			f := NewFileManager(2)
			scenario.test(f.Format(scenario.filename))
		})
	}
}

func TestFileManagerFormatAndReplace(t *testing.T) {
	type scenario struct {
		testName string
		path     string
		setup    func()
		test     func([]interface{})
	}

	scenarios := []scenario{
		{
			"format a file",
			"tmp/file1.feature",
			func() {
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
			},
			func(output []interface{}) {
				assertNoErrors(t, output)

				content := `Feature: test
  test

  Scenario: scenario1
    Given whatever
    Then whatever
      """
      hello world
      """
`

				b, e := os.ReadFile("tmp/file1.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, content, string(b))
			},
		},
		{
			"compact json in example",
			"tmp/file1.feature",
			func() {
				content := []byte(`Feature: test feature

Scenario Outline: Compact json
Given I load data:
  """
  <data>
  """
Examples: 
  | data                              |
  |{"key1": "value2",    "key2": "value2"}|
  |[1,    2,   3]                         |
`)

				assert.NoError(t, os.RemoveAll("tmp/"))
				assert.NoError(t, os.MkdirAll("tmp/", 0o777))
				assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o600))
			},
			func(output []interface{}) {
				assertNoErrors(t, output)

				content := `Feature: test feature

  Scenario Outline: Compact json
    Given I load data:
      """
      <data>
      """
    Examples:
      | data                              |
      | {"key1":"value2","key2":"value2"} |
      | [1,2,3]                           |
`

				b, e := os.ReadFile("tmp/file1.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, content, string(b))
			},
		},
		{
			"format bullet points",
			"tmp/file1.feature",
			func() {
				content := []byte(`Feature: bullet points

Scenario: format bullet points
Given Some state
* Another state
* Yet another state
When check formatting
Then all is good
`)

				assert.NoError(t, os.RemoveAll("tmp/"))
				assert.NoError(t, os.MkdirAll("tmp/", 0o777))
				assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o600))
			},
			func(output []interface{}) {
				assertNoErrors(t, output)

				content := `Feature: bullet points

  Scenario: format bullet points
    Given Some state
    * Another state
    * Yet another state
    When check formatting
    Then all is good
`

				b, e := os.ReadFile("tmp/file1.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, content, string(b))
			},
		},
		{
			"format a folder",
			"tmp/",
			func() {
				content := []byte(`Feature: test
  test

Scenario:   scenario%d
  Given           whatever
  Then      whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("tmp/"))
				assert.NoError(t, os.MkdirAll("tmp/", 0o777))
				assert.NoError(t, os.MkdirAll("tmp/test1", 0o777))
				assert.NoError(t, os.MkdirAll("tmp/test2/test3", 0o777))

				for i, f := range []string{
					"tmp/file1.feature",
					"tmp/file2.feature",
					"tmp/test1/file3.feature",
					"tmp/test1/file4.feature",
					"tmp/test2/test3/file5.feature",
					"tmp/test2/test3/file6.feature",
				} {
					assert.NoError(t, os.WriteFile(f, []byte(fmt.Sprintf(string(content), i)), 0o600))
				}
			},
			func(output []interface{}) {
				assertNoErrors(t, output)

				content := `Feature: test
  test

  Scenario: scenario%d
    Given whatever
    Then whatever
      """
      hello world
      """
`

				for i, f := range []string{
					"tmp/file1.feature",
					"tmp/file2.feature",
					"tmp/test1/file3.feature",
					"tmp/test1/file4.feature",
					"tmp/test2/test3/file5.feature",
					"tmp/test2/test3/file6.feature",
				} {
					b, e := os.ReadFile(f)
					assert.NoError(t, e)
					assert.EqualValues(t, fmt.Sprintf(content, i), string(b))
				}
			},
		},
		{
			"format a folder with parsing errors",
			"tmp/",
			func() {
				content := []byte(`Feature: test
      test

Scenario:   scenario
   Given           whatever
   Then      whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("tmp"))
				assert.NoError(t, os.MkdirAll("tmp", 0o777))
				assert.NoError(t, os.MkdirAll("tmp/test1", 0o777))

				assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o600))
				assert.NoError(t, os.WriteFile("tmp/file2.feature", append([]byte("whatever"), content...), 0o600))
				assert.NoError(t, os.WriteFile("tmp/test1/file3.feature", content, 0o600))
				assert.NoError(t, os.WriteFile("tmp/test1/file4.feature", content, 0o600))
				assert.NoError(t, os.WriteFile("tmp/test1/file5.feature", append([]byte("something"), content...), 0o600))
			},
			func(output []interface{}) {
				assert.Len(t, output, 5)

				expectedErrs := []string{
					`an error occurred with file "tmp/file2.feature" : Parser errors:
(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whateverFeature: test'`,
					`an error occurred with file "tmp/test1/file5.feature" : Parser errors:
(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'somethingFeature: test'`,
				}

				i := 0
				for _, expectedErr := range expectedErrs {
					for _, o := range output {
						e, ok := o.(error)
						if ok && expectedErr == e.Error() {
							i++
						}
					}
				}
				if l := len(expectedErrs); i != l {
					assert.Fail(t, fmt.Sprintf("Must fail with %v files when formatting folder", l))
				}
			},
		},
		{
			"format folder with no feature files",
			"tmp/",
			func() {
				assert.NoError(t, os.RemoveAll("tmp/"))
				assert.NoError(t, os.MkdirAll("tmp/", 0o777))
				assert.NoError(t, os.WriteFile("tmp/file1.txt", []byte("file1"), 0o600))
				assert.NoError(t, os.WriteFile("tmp/file2.txt", []byte("file2"), 0o600))
			},
			func(output []interface{}) {
				assertNoErrors(t, output)
			},
		},
		{
			"format an unexisting folder",
			"whatever/whatever",
			func() {},
			func(output []interface{}) {
				assert.Len(t, output, 1)
				e, _ := output[0].(error)
				assert.Error(t, e)
				assert.EqualError(t, e, "stat whatever/whatever: no such file or directory")
			},
		},
		{
			"format an invalid file",
			"features/invalid.feature",
			func() {},
			func(output []interface{}) {
				assert.Len(t, output, 1)
				e, _ := output[0].(error)
				assert.Error(t, e)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.testName, func(_ *testing.T) {
			scenario.setup()
			f := NewFileManager(2)
			scenario.test(f.FormatAndReplace(scenario.path))
			// Cleanup
			_ = os.RemoveAll("tmp/")
		})
	}
}

func TestFileManagerCheck(t *testing.T) {
	type scenario struct {
		testName string
		path     string
		setup    func()
		test     func([]interface{})
	}

	scenarios := []scenario{
		{
			"Check a file wrongly formatted",
			"tmp/file1.feature",
			func() {
				content := []byte(`Feature: test
   test

Scenario:            scenario1
   Given       whatever
   Then                  whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("tmp"))
				assert.NoError(t, os.MkdirAll("tmp", 0o777))
				assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o600))
			},
			func(output []interface{}) {
				assert.Len(t, output, 1)
				e, _ := output[0].(error)
				assert.Error(t, e)
				assert.EqualError(t, e, `an error occurred with file "tmp/file1.feature" : file is not properly formatted`)
			},
		},
		{
			"Check a file correctly formatted",
			"tmp/file1.feature",
			func() {
				content := []byte(`Feature: test

  Scenario: scenario
    Given whatever
    Then whatever
      """
      hello world
      """
`)

				assert.NoError(t, os.RemoveAll("tmp"))
				assert.NoError(t, os.MkdirAll("tmp", 0o777))
				assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o600))
			},
			func(output []interface{}) {
				assertNoErrors(t, output)
			},
		},
		{
			"Check a folder is wrongly formatted",
			"tmp/",
			func() {
				content := []byte(`Feature: test
   test

Scenario:   scenario%d
   Given           whatever
   Then      whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("tmp"))
				assert.NoError(t, os.MkdirAll("tmp", 0o777))
				assert.NoError(t, os.MkdirAll("tmp/test1", 0o777))
				assert.NoError(t, os.MkdirAll("tmp/test2/test3", 0o777))

				for i, f := range []string{
					"tmp/file1.feature",
					"tmp/file2.feature",
					"tmp/test1/file3.feature",
					"tmp/test1/file4.feature",
					"tmp/test2/test3/file5.feature",
					"tmp/test2/test3/file6.feature",
				} {
					assert.NoError(t, os.WriteFile(f, []byte(fmt.Sprintf(string(content), i)), 0o600))
				}
			},
			func(output []interface{}) {
				assert.Len(t, output, 6)

				expectedErrs := []string{
					`an error occurred with file "tmp/file1.feature" : file is not properly formatted`,
					`an error occurred with file "tmp/file2.feature" : file is not properly formatted`,
					`an error occurred with file "tmp/test1/file3.feature" : file is not properly formatted`,
					`an error occurred with file "tmp/test1/file4.feature" : file is not properly formatted`,
					`an error occurred with file "tmp/test2/test3/file5.feature" : file is not properly formatted`,
					`an error occurred with file "tmp/test2/test3/file6.feature" : file is not properly formatted`,
				}

				i := 0
				for _, expectedErr := range expectedErrs {
					for _, o := range output {
						e, ok := o.(error)
						if ok && expectedErr == e.Error() {
							i++
						}
					}
				}
				if l := len(expectedErrs); i != l {
					assert.Fail(t, fmt.Sprintf("Must fail with %v files when formatting folder", l))
				}
			},
		},
		{
			"Check a folder is correctly formatted",
			"tmp/",
			func() {
				content := []byte(`Feature: test

  Scenario: scenario%d
    Given whatever
    Then whatever
      """
      hello world
      """
`)

				assert.NoError(t, os.RemoveAll("tmp"))
				assert.NoError(t, os.MkdirAll("tmp", 0o777))
				assert.NoError(t, os.MkdirAll("tmp/test1", 0o777))
				assert.NoError(t, os.MkdirAll("tmp/test2/test3", 0o777))

				for i, f := range []string{
					"tmp/file1.feature",
					"tmp/file2.feature",
					"tmp/test1/file3.feature",
					"tmp/test1/file4.feature",
					"tmp/test2/test3/file5.feature",
					"tmp/test2/test3/file6.feature",
				} {
					assert.NoError(t, os.WriteFile(f, []byte(fmt.Sprintf(string(content), i)), 0o600))
				}
			},
			func(output []interface{}) {
				assertNoErrors(t, output)
			},
		},
		{
			"Check a folder with parsing errors",
			"tmp/",
			func() {
				content := []byte(`Feature: test

  Scenario: scenario
    Given whatever
    Then whatever
      """
      hello world
      """
`)

				assert.NoError(t, os.RemoveAll("tmp"))
				assert.NoError(t, os.MkdirAll("tmp", 0o777))
				assert.NoError(t, os.MkdirAll("tmp/test1", 0o777))

				assert.NoError(t, os.WriteFile("tmp/file1.feature", content, 0o600))
				assert.NoError(t, os.WriteFile("tmp/file2.feature", append([]byte("whatever"), content...), 0o600))
				assert.NoError(t, os.WriteFile("tmp/test1/file3.feature", content, 0o600))
				assert.NoError(t, os.WriteFile("tmp/test1/file4.feature", content, 0o600))
				assert.NoError(t, os.WriteFile("tmp/test1/file5.feature", append([]byte("something"), content...), 0o600))
			},
			func(output []interface{}) {
				assert.Len(t, output, 5)

				expectedErrs := []string{
					`an error occurred with file "tmp/file2.feature" : Parser errors:
(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whateverFeature: test'`,
					`an error occurred with file "tmp/test1/file5.feature" : Parser errors:
(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'somethingFeature: test'`,
				}

				i := 0
				for _, expectedErr := range expectedErrs {
					for _, o := range output {
						e, ok := o.(error)
						if ok && expectedErr == e.Error() {
							i++
						}
					}
				}
				if l := len(expectedErrs); i != l {
					assert.Fail(t, fmt.Sprintf("Must fail with %v files when formatting folder", l))
				}
			},
		},
		{
			"Check folder with no feature files",
			"tmp",
			func() {
				assert.NoError(t, os.RemoveAll("tmp"))
				assert.NoError(t, os.MkdirAll("tmp", 0o777))
				assert.NoError(t, os.WriteFile("tmp/file1.txt", []byte("file1"), 0o600))
				assert.NoError(t, os.WriteFile("tmp/file2.txt", []byte("file2"), 0o600))
			},
			func(output []interface{}) {
				assertNoErrors(t, output)
			},
		},
		{
			"Check an unexisting folder",
			"whatever/whatever",
			func() {},
			func(output []interface{}) {
				assert.Len(t, output, 1)
				e, _ := output[0].(error)
				assert.Error(t, e)
				assert.EqualError(t, e, "stat whatever/whatever: no such file or directory")
			},
		},
		{
			"Check an invalid file",
			"features/invalid.feature",
			func() {},
			func(output []interface{}) {
				assert.Len(t, output, 1)
				e, _ := output[0].(error)
				assert.Error(t, e)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.testName, func(_ *testing.T) {
			scenario.setup()

			f := NewFileManager(2)

			scenario.test(f.Check(scenario.path))
			// Cleanup
			_ = os.RemoveAll("tmp/")
		})
	}
}

func assertNoErrors(t *testing.T, any []interface{}) {
	for _, a := range any {
		if _, ok := a.(error); ok {
			assert.Fail(t, "An error is not expected.", a)
		}
	}
}
