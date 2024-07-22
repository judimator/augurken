package json

import (
	"strconv"
	"sync"
)

// Valid reports whether data is a valid JSON encoding.
func Valid(data []byte) bool {
	scan := newScanner()
	defer freeScanner(scan)

	return checkValid(data, scan) == nil
}

func checkValid(data []byte, scan *scanner) error {
	scan.reset()

	for _, c := range data {
		scan.bytes++
		if scan.step(scan, c) == scanError {
			return scan.err
		}
	}

	if scan.eof() == scanError {
		return scan.err
	}

	return nil
}

type SyntaxError struct {
	msg    string
	Offset int64
}

func (e *SyntaxError) Error() string { return e.msg }

type scanner struct {
	step             func(*scanner, byte) int
	endTop           bool
	parseState       []int
	err              error
	bytes            int64
	placeholderStack placeholderStack
}

var scannerPool = sync.Pool{
	New: func() any {
		return &scanner{}
	},
}

type placeholderStack struct {
	chars  []byte
	length int
}

func (s *placeholderStack) push(char byte) {
	s.chars = append(s.chars, char)
	s.length++
}

func (s *placeholderStack) pop() byte {
	char := s.chars[len(s.chars)-1]

	s.chars = s.chars[0 : len(s.chars)-1]
	s.length--

	return char
}

func newScanner() *scanner {
	scan := scannerPool.Get().(*scanner)
	// scan.reset by design doesn't set bytes to zero
	scan.bytes = 0
	scan.placeholderStack = placeholderStack{}
	scan.reset()

	return scan
}

func freeScanner(scan *scanner) {
	// Avoid hanging on to too much memory in extreme cases.
	if len(scan.parseState) > 1024 {
		scan.parseState = nil
	}

	scannerPool.Put(scan)
}

const (
	// Continue.
	scanContinue                  = iota // uninteresting byte
	scanBeginLiteral                     // end implied by next result != scanContinue
	scanBeginObject                      // begin object
	scanObjectKey                        // just finished object key (string)
	scanObjectValue                      // just finished non-last object value
	scanEndObject                        // end object (implies scanObjectValue if possible)
	scanBeginArray                       // begin array
	scanArrayValue                       // just finished array value
	scanEndArray                         // end array (implies scanArrayValue if possible)
	scanSkipSpace                        // space byte; can skip; known to be last "continue" result
	scanBeginPlaceholder                 // begin of placeholder value, i.e. `<`
	scanEndPlaceholder                   // end of placeholder value, i.e. `>`
	scanContinueAfterMissingComma        // force to continue if no comma after prev element

	// Stop.
	scanEnd   // top-level value ended *before* this byte; known to be first "stop" result
	scanError // hit an error, scanner.err.
)

// These values are stored in the parseState stack.
// They give the current state of a composite value
// being scanned. If the parser is inside a nested value
// the parseState describes the nested state, outermost at entry 0.
const (
	parseObjectKey   = iota // parsing object key (before colon)
	parseObjectValue        // parsing object value (after colon)
	parseArrayValue         // parsing array value
)

// This limits the max nesting depth to prevent stack overflow.
// This is permitted by https://tools.ietf.org/html/rfc7159#section-9
const maxNestingDepth = 10000

// reset prepares the scanner for use.
// It must be called before calling s.step.
func (s *scanner) reset() {
	s.step = stateBeginValue
	s.parseState = s.parseState[0:0]
	s.placeholderStack.chars = s.placeholderStack.chars[0:0]
	s.err = nil
	s.endTop = false
}

// eof tells the scanner that the end of input has been reached.
// It returns a scan status just as s.step does.
func (s *scanner) eof() int {
	if s.err != nil {
		return scanError
	}

	if s.endTop {
		return scanEnd
	}

	s.step(s, ' ')

	if s.endTop {
		return scanEnd
	}

	if s.err == nil {
		s.err = &SyntaxError{"unexpected end of JSON input", s.bytes}
	}

	return scanError
}

// pushParseState pushes a new parse state p onto the parse stack.
// an error state is returned if maxNestingDepth was exceeded, otherwise successState is returned.
func (s *scanner) pushParseState(c byte, newParseState int, successState int) int {
	s.parseState = append(s.parseState, newParseState)
	if len(s.parseState) <= maxNestingDepth {
		return successState
	}

	return s.error(c, "exceeded max JSON depth")
}

// popParseState pops a parse state (already obtained) off the stack
// and updates s.step accordingly.
func (s *scanner) popParseState() {
	n := len(s.parseState) - 1
	s.parseState = s.parseState[0:n]

	if n == 0 {
		s.step = stateEndTop
		s.endTop = true
	} else {
		s.step = stateEndValue
	}
}

func isSpace(c byte) bool {
	return c <= ' ' && (c == ' ' || c == '\t' || c == '\r' || c == '\n')
}

// stateBeginValueOrEmpty is the state after reading `[`.
func stateBeginValueOrEmpty(s *scanner, c byte) int {
	if isSpace(c) {
		return scanSkipSpace
	}

	if c == ']' {
		return stateEndValue(s, c)
	}

	return stateBeginValue(s, c)
}

// stateBeginValue is the state at the beginning of the input.
func stateBeginValue(s *scanner, c byte) int {
	if isSpace(c) {
		return scanSkipSpace
	}

	switch c {
	case '{':
		s.step = stateBeginStringOrEmpty

		return s.pushParseState(c, parseObjectKey, scanBeginObject)
	case '[':
		s.step = stateBeginValueOrEmpty

		return s.pushParseState(c, parseArrayValue, scanBeginArray)
	case '"':
		s.step = stateInString

		return scanBeginLiteral
	case '-':
		s.step = stateNeg

		return scanBeginLiteral
	case '0': // Beginning of 0.123
		s.step = state0

		return scanBeginLiteral
	case 't': // Beginning of true
		s.step = stateT

		return scanBeginLiteral
	case 'f': // Beginning of false
		s.step = stateF

		return scanBeginLiteral
	case 'n': // Beginning of null
		s.step = stateN

		return scanBeginLiteral
	case '<':
		s.step = stateInPlaceholder
		s.placeholderStack.push(c)

		return scanBeginPlaceholder
	}

	if '1' <= c && c <= '9' { // Beginning of 1234.5
		s.step = state1

		return scanBeginLiteral
	}

	return s.error(c, "looking for beginning of value")
}

// stateBeginStringOrEmpty is the state after reading `{`.
func stateBeginStringOrEmpty(s *scanner, c byte) int {
	if isSpace(c) {
		return scanSkipSpace
	}

	if c == '}' {
		n := len(s.parseState)
		s.parseState[n-1] = parseObjectValue

		return stateEndValue(s, c)
	}

	return stateBeginStringOrPlaceHolder(s, c)
}

// stateInPlaceholder is the state after reading <placeholder>
func stateInPlaceholder(s *scanner, c byte) int {
	switch c {
	case '>':
		if s.placeholderStack.length == 0 {
			return s.error(c, "Invalid placeholder given")
		}

		s.placeholderStack.pop()

		if s.placeholderStack.length >= 1 {
			return scanContinue
		}

		n := len(s.parseState)
		if n == 0 {
			// Completed top-level before the current byte.
			s.step = stateEndTop
			s.endTop = true

			return stateEndTop(s, c)
		}
		// Guess it is `{...,<placeholder>,...}`. Consider that `parseObjectKey` finished
		if s.parseState[n-1] == parseObjectKey {
			s.parseState[n-1] = parseObjectValue
		}

		s.step = stateEndValue

		return scanEndPlaceholder
	case '<':
		s.placeholderStack.push(c)

		return scanContinue
	default:
		return scanContinue
	}
}

// stateBeginStringOrPlaceHolder is the state after reading `{"key": value,`.
func stateBeginStringOrPlaceHolder(s *scanner, c byte) int {
	if isSpace(c) {
		return scanSkipSpace
	}

	if c == '"' {
		s.step = stateInString

		return scanBeginLiteral
	}

	if c == '<' {
		s.step = stateInPlaceholder
		s.placeholderStack.push(c)

		return scanBeginPlaceholder
	}

	return s.error(c, "looking for beginning of object key string or placeholder")
}

// stateEndValue is the state after completing a value,
// such as after reading `{}` or `true` or `["x"`.
func stateEndValue(s *scanner, c byte) int {
	n := len(s.parseState)
	if n == 0 {
		// Completed top-level before the current byte.
		s.step = stateEndTop
		s.endTop = true

		return stateEndTop(s, c)
	}

	if isSpace(c) {
		s.step = stateEndValue

		return scanSkipSpace
	}

	ps := s.parseState[n-1]
	switch ps {
	case parseObjectKey:
		if c == ':' {
			s.parseState[n-1] = parseObjectValue
			s.step = stateBeginValue

			return scanObjectKey
		}

		return s.error(c, "after object key")
	case parseObjectValue:
		if c == ',' {
			s.parseState[n-1] = parseObjectKey
			s.step = stateBeginValue

			return scanObjectValue
		}

		if c == '}' {
			s.popParseState()

			return scanEndObject
		}
		// Reading after `{<placeholder>}`. It is the case when no comma after `{<placeholder>}`.
		// Next might be either `"key": {` or `<placeholder>`
		if c == '"' || c == '<' {
			s.parseState[n-1] = parseObjectKey

			return stateInStringAfterMissingComma(s, c)
		}

		return s.error(c, "after object key:value pair or placeholder")
	case parseArrayValue:
		if c == ',' {
			s.step = stateBeginValue

			return scanArrayValue
		}
		// Next might be `<placeholder>`
		if c == '<' {
			return stateInStringAfterMissingComma(s, c)
		}

		if c == ']' {
			s.popParseState()

			return scanEndArray
		}

		return s.error(c, "after array element")
	}

	return s.error(c, "")
}

// stateEndTop is the state after finishing the top-level value,
// such as after reading `{}` or `[1,2,3]`.
// Only space characters should be seen now.
func stateEndTop(s *scanner, c byte) int {
	if !isSpace(c) {
		// Complain about non-space byte on next call.
		s.error(c, "after top-level value")
	}

	return scanEnd
}

// stateInString is the state after reading `"`.
func stateInString(s *scanner, c byte) int {
	if c == '"' {
		s.step = stateEndValue

		return scanContinue
	}

	if c == '\\' {
		s.step = stateInStringEsc

		return scanContinue
	}

	if c < 0x20 {
		return s.error(c, "in string literal")
	}

	return scanContinue
}

func stateInStringAfterMissingComma(s *scanner, c byte) int {
	if c == '"' {
		s.step = stateInString

		return scanContinueAfterMissingComma
	}

	if c == '<' {
		s.step = stateInPlaceholder
		s.placeholderStack.push(c)

		return scanContinueAfterMissingComma
	}

	return scanContinue
}

// stateInStringEsc is the state after reading `"\` during a quoted string.
func stateInStringEsc(s *scanner, c byte) int {
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		s.step = stateInString

		return scanContinue
	case 'u':
		s.step = stateInStringEscU

		return scanContinue
	}

	return s.error(c, "in string escape code")
}

// stateInStringEscU is the state after reading `"\u` during a quoted string.
func stateInStringEscU(s *scanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU1

		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU1 is the state after reading `"\u1` during a quoted string.
func stateInStringEscU1(s *scanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU12

		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU12 is the state after reading `"\u12` during a quoted string.
func stateInStringEscU12(s *scanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU123

		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU123 is the state after reading `"\u123` during a quoted string.
func stateInStringEscU123(s *scanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInString

		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateNeg is the state after reading `-` during a number.
func stateNeg(s *scanner, c byte) int {
	if c == '0' {
		s.step = state0

		return scanContinue
	}

	if '1' <= c && c <= '9' {
		s.step = state1

		return scanContinue
	}

	return s.error(c, "in numeric literal")
}

// state1 is the state after reading a non-zero integer during a number,
// such as after reading `1` or `100` but not `0`.
func state1(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		s.step = state1

		return scanContinue
	}

	return state0(s, c)
}

// state0 is the state after reading `0` during a number.
func state0(s *scanner, c byte) int {
	if c == '.' {
		s.step = stateDot

		return scanContinue
	}

	if c == 'e' || c == 'E' {
		s.step = stateE

		return scanContinue
	}

	return stateEndValue(s, c)
}

// stateDot is the state after reading the integer and decimal point in a number,
// such as after reading `1.`.
func stateDot(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		s.step = stateDot0

		return scanContinue
	}

	return s.error(c, "after decimal point in numeric literal")
}

// stateDot0 is the state after reading the integer, decimal point, and subsequent
// digits of a number, such as after reading `3.14`.
func stateDot0(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		return scanContinue
	}

	if c == 'e' || c == 'E' {
		s.step = stateE

		return scanContinue
	}

	return stateEndValue(s, c)
}

// stateE is the state after reading the mantissa and e in a number,
// such as after reading `314e` or `0.314e`.
func stateE(s *scanner, c byte) int {
	if c == '+' || c == '-' {
		s.step = stateESign

		return scanContinue
	}

	return stateESign(s, c)
}

// stateESign is the state after reading the mantissa, e, and sign in a number,
// such as after reading `314e-` or `0.314e+`.
func stateESign(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		s.step = stateE0

		return scanContinue
	}

	return s.error(c, "in exponent of numeric literal")
}

// stateE0 is the state after reading the mantissa, e, optional sign,
// and at least one digit of the exponent in a number,
// such as after reading `314e-2` or `0.314e+1` or `3.14e0`.
func stateE0(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		return scanContinue
	}

	return stateEndValue(s, c)
}

// stateT is the state after reading `t`.
func stateT(s *scanner, c byte) int {
	if c == 'r' {
		s.step = stateTr

		return scanContinue
	}

	return s.error(c, "in literal true (expecting 'r')")
}

// stateTr is the state after reading `tr`.
func stateTr(s *scanner, c byte) int {
	if c == 'u' {
		s.step = stateTru

		return scanContinue
	}

	return s.error(c, "in literal true (expecting 'u')")
}

// stateTru is the state after reading `tru`.
func stateTru(s *scanner, c byte) int {
	if c == 'e' {
		s.step = stateEndValue

		return scanContinue
	}

	return s.error(c, "in literal true (expecting 'e')")
}

// stateF is the state after reading `f`.
func stateF(s *scanner, c byte) int {
	if c == 'a' {
		s.step = stateFa

		return scanContinue
	}

	return s.error(c, "in literal false (expecting 'a')")
}

// stateFa is the state after reading `fa`.
func stateFa(s *scanner, c byte) int {
	if c == 'l' {
		s.step = stateFal

		return scanContinue
	}

	return s.error(c, "in literal false (expecting 'l')")
}

// stateFal is the state after reading `fal`.
func stateFal(s *scanner, c byte) int {
	if c == 's' {
		s.step = stateFals

		return scanContinue
	}

	return s.error(c, "in literal false (expecting 's')")
}

// stateFals is the state after reading `fals`.
func stateFals(s *scanner, c byte) int {
	if c == 'e' {
		s.step = stateEndValue

		return scanContinue
	}

	return s.error(c, "in literal false (expecting 'e')")
}

// stateN is the state after reading `n`.
func stateN(s *scanner, c byte) int {
	if c == 'u' {
		s.step = stateNu

		return scanContinue
	}

	return s.error(c, "in literal null (expecting 'u')")
}

// stateNu is the state after reading `nu`.
func stateNu(s *scanner, c byte) int {
	if c == 'l' {
		s.step = stateNul

		return scanContinue
	}

	return s.error(c, "in literal null (expecting 'l')")
}

// stateNul is the state after reading `nul`.
func stateNul(s *scanner, c byte) int {
	if c == 'l' {
		s.step = stateEndValue

		return scanContinue
	}

	return s.error(c, "in literal null (expecting 'l')")
}

// stateError is the state after reaching a syntax error,
// such as after reading `[1}` or `5.1.2`.
func stateError(_ *scanner, _ byte) int {
	return scanError
}

// error records an error and switches to the error state.
func (s *scanner) error(c byte, context string) int {
	s.step = stateError
	s.err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, s.bytes}

	return scanError
}

// quoteChar formats c as a quoted character literal.
func quoteChar(c byte) string {
	// special cases - different from quoted strings
	if c == '\'' {
		return `'\''`
	}

	if c == '"' {
		return `'"'`
	}
	// use quoted string with different quotation marks
	s := strconv.Quote(string(c))

	return "'" + s[1:len(s)-1] + "'"
}
