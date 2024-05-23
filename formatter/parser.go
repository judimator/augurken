package formatter

import (
	"bytes"

	gherkin "github.com/cucumber/gherkin/go/v28"
)

func parse(content []byte) (*token, error) {
	token := &token{}
	builder := &node{token: token}
	matcher := gherkin.NewMatcher(gherkin.DialectsBuiltin())
	scanner := gherkin.NewScanner(bytes.NewBuffer(content))
	parser := gherkin.NewParser(builder)
	parser.StopAtFirstError(true)

	return token, parser.Parse(scanner, matcher)
}
