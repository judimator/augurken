package formatter

import (
	"errors"

	gherkin "github.com/cucumber/gherkin/go/v28"
)

type node struct {
	token *token
}

func (t *node) Build(tok *gherkin.Token) (bool, error) {
	if tok == nil {
		return false, errors.New("token is not defined")
	}

	if tok.IsEOF() {
		return true, nil
	}

	switch {
	case t.token == nil:
		t.token = &token{kind: tok.Type, values: []*gherkin.Token{}}
	case tok.Type != t.token.kind:
		t.token.nex = &token{kind: tok.Type, values: []*gherkin.Token{}, prev: t.token}
		t.token = t.token.nex
	}

	t.token.values = append(t.token.values, tok)

	return true, nil
}

func (t *node) StartRule(_ gherkin.RuleType) (bool, error) {
	return true, nil
}

func (t *node) EndRule(_ gherkin.RuleType) (bool, error) {
	return true, nil
}

func (t *node) Reset() {
}

type token struct {
	kind   gherkin.TokenType
	values []*gherkin.Token
	prev   *token
	nex    *token
}

func (t *token) isExcluded(kind gherkin.TokenType, excluded []gherkin.TokenType) bool {
	for _, e := range excluded {
		if kind == e {
			return true
		}
	}

	return false
}

func (t *token) previous(excluded []gherkin.TokenType) *token {
	for tok := t.prev; tok != nil; tok = tok.prev {
		if !t.isExcluded(tok.kind, excluded) {
			return tok
		}
	}

	return nil
}

func (t *token) next(excluded []gherkin.TokenType) *token {
	for tok := t.nex; tok != nil; tok = tok.nex {
		if !t.isExcluded(tok.kind, excluded) {
			return tok
		}
	}

	return nil
}
