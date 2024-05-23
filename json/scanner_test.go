package json

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type CasePos struct{ pc [1]uintptr }

type CaseName struct {
	Name  string
	Where CasePos
}

func Name(s string) (c CaseName) {
	c.Name = s
	runtime.Callers(2, c.Where.pc[:])

	return c
}

func (pos CasePos) String() string {
	frames := runtime.CallersFrames(pos.pc[:])
	frame, _ := frames.Next()

	return fmt.Sprintf("%s:%d", path.Base(frame.File), frame.Line)
}

func indentNewlines(s string) string {
	return strings.Join(strings.Split(s, "\n"), "\n\t")
}

func TestValid(t *testing.T) {
	tests := []struct {
		CaseName
		data string
		ok   bool
	}{
		{Name(""), `foo`, false},
		{Name(""), `}{`, false},
		{Name(""), `{]`, false},
		{Name(""), `{}`, true},
		{Name(""), `{"foo":"bar"}`, true},
		{Name(""), `{"foo":"bar","bar":{"baz":["qux"]}}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if ok := Valid([]byte(tt.data)); ok != tt.ok {
				t.Errorf("%s: Valid(`%s`) = %v, want %v", tt.Where, tt.data, ok, tt.ok)
			}
		})
	}
}

func TestIndent(t *testing.T) {
	tests := []struct {
		CaseName
		compact string
		indent  string
	}{
		{Name(""), `1`, `1`},
		{Name(""), `{}`, `{}`},
		{Name(""), `[]`, `[]`},
		{Name(""), `{"":2}`, "{\n    \"\": 2\n}"},
		{Name(""), `[3]`, "[\n    3\n]"},
		{Name(""), `[1,2,3]`, "[\n    1,\n    2,\n    3\n]"},
		{Name(""), `{"x":1}`, "{\n    \"x\": 1\n}"},
		{Name(""), `[true,false,null,"x",1,1.5,0,-5e+2]`, `[
    true,
    false,
    null,
    "x",
    1,
    1.5,
    0,
    -5e+2
]`},
		{Name(""), `{"x":1, "y":<any>, <p1>,<p2>,"z":{<p3>,<p4>,"key": <p5>}}`, `{
    "x": 1,
    "y": <any>,
    <p1>,
    <p2>,
    "z": {
        <p3>,
        <p4>,
        "key": <p5>
    }
}`},
		{Name(""), "{\"\":\"<>&\u2028\u2029\"}", "{\n    \"\": \"<>&\u2028\u2029\"\n}"}, // See golang.org/issue/34070
	}

	var buf bytes.Buffer

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			buf.Reset()
			// 4 spaces as indention
			if err := Indent(&buf, []byte(tt.compact), "", "    "); err != nil {
				t.Errorf("%s: Indent error: %v", tt.Where, err)
			} else if got := buf.String(); got != tt.indent {
				t.Errorf("%s: Indent:\n\tgot:  %s\n\twant: %s", tt.Where, indentNewlines(got), indentNewlines(tt.indent))
			}
		})
	}
}

func TestIndentErrors(t *testing.T) {
	tests := []struct {
		CaseName
		in  string
		err error
	}{
		{Name(""), `{"X": "foo", "Y"}`, &SyntaxError{"invalid character '}' after object key", 17}},
		{Name(""), `{"X": "foo" "Y": "bar"}`, &SyntaxError{"invalid character '\"' after object key:value pair", 13}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			slice := make([]uint8, 0)
			buf := bytes.NewBuffer(slice)

			if err := Indent(buf, []uint8(tt.in), "", ""); err != nil {
				if !reflect.DeepEqual(err, tt.err) {
					t.Fatalf("%s: Indent error:\n\tgot:  %v\n\twant: %v", tt.Where, err, tt.err)
				}
			}
		})
	}
}
