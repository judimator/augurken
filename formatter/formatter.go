package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	gherkin "github.com/cucumber/gherkin/go/v28"
	augurkenjson "github.com/judimator/augurken/json"
)

func format(token *token, indent int) ([]byte, error) {
	paddings := map[gherkin.TokenType]int{
		gherkin.TokenTypeFeatureLine:        0,
		gherkin.TokenTypeBackgroundLine:     indent,
		gherkin.TokenTypeScenarioLine:       indent,
		gherkin.TokenTypeDocStringSeparator: 3 * indent,
		gherkin.TokenTypeStepLine:           2 * indent,
		gherkin.TokenTypeExamplesLine:       2 * indent,
		gherkin.TokenTypeOther:              3 * indent,
		gherkin.TokenTypeTableRow:           3 * indent,
	}

	formats := map[gherkin.TokenType]func(values []*gherkin.Token) []string{
		gherkin.TokenTypeFeatureLine:        extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeBackgroundLine:     extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeScenarioLine:       extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeExamplesLine:       extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeComment:            extractTokensText,
		gherkin.TokenTypeTagLine:            extractTokensItemsText,
		gherkin.TokenTypeDocStringSeparator: extractKeyword,
		gherkin.TokenTypeRuleLine:           extractKeywordAndTextSeparatedWithAColon,
		gherkin.TokenTypeOther:              extractTokensText,
		gherkin.TokenTypeStepLine:           extractTokensKeywordAndText,
		gherkin.TokenTypeTableRow:           extractTableRowsAndComments,
		gherkin.TokenTypeEmpty:              extractTokensItemsText,
	}

	var document []string
	optionalRulePadding := 0
	var accumulator []*gherkin.Token

	for tok := token; tok != nil; tok = tok.nex {
		values := tok.values

		if len(accumulator) > 0 &&
			tok.kind == gherkin.TokenTypeTableRow &&
			(tok.nex != nil && tok.nex.kind != gherkin.TokenTypeComment) || tok.nex == nil {
			values = append(accumulator, tok.values...)
			accumulator = []*gherkin.Token{}
		}
		if tok.kind == gherkin.TokenTypeTableRow &&
			tok.nex != nil &&
			tok.nex.kind == gherkin.TokenTypeComment &&
			tok.nex.nex != nil &&
			tok.nex.nex.kind == gherkin.TokenTypeTableRow ||
			len(accumulator) > 0 && tok.kind == gherkin.TokenTypeComment ||
			len(accumulator) > 0 && tok.kind == gherkin.TokenTypeTableRow {
			accumulator = append(accumulator, tok.values...)
			continue
		}
		if tok.kind == 0 {
			continue
		}

		padding := paddings[tok.kind] + optionalRulePadding
		lines := formats[tok.kind](values)

		if tok.kind == gherkin.TokenTypeRuleLine {
			optionalRulePadding = indent
			padding = indent
		} else if tok.kind == gherkin.TokenTypeComment {
			padding = getTagOrCommentPadding(paddings, indent, tok)
			lines = trimLinesSpace(lines)
		} else if tok.kind == gherkin.TokenTypeTagLine {
			padding = getTagOrCommentPadding(paddings, indent, tok)
		} else if tok.kind == gherkin.TokenTypeDocStringSeparator {
			lines = extractKeyword(tok.values)
		} else if tok.kind == gherkin.TokenTypeOther {
			if isDescriptionFeature(tok) {
				padding = indent
			} else if tok.prev.kind == gherkin.TokenTypeDocStringSeparator &&
				tok.nex.kind == gherkin.TokenTypeDocStringSeparator {
				var buffer bytes.Buffer

				// Transform into string and get bytes
				source := []byte(strings.Join(lines, " "))
				prefixSpace := strings.Repeat(" ", padding)
				indentSpace := strings.Repeat(" ", indent)

				if ok := augurkenjson.Valid(source); ok == true {
					_ = augurkenjson.Indent(&buffer, source, prefixSpace, indentSpace)
					lines = []string{string(buffer.Bytes())}
				}
				// TODO: Handle json error and print col and line
			}

			lines = trimLinesSpace(lines)
		}

		document = append(document, trimExtraTrailingSpace(indentStrings(padding, lines))...)
	}
	return []byte(strings.Join(document, "\n") + "\n"), nil
}

func getTagOrCommentPadding(paddings map[gherkin.TokenType]int, indent int, tok *token) int {
	var kind gherkin.TokenType
	excluded := []gherkin.TokenType{gherkin.TokenTypeTagLine, gherkin.TokenTypeComment}
	if tok.next(excluded) != nil {
		if s := tok.next(excluded); s != nil {
			kind = s.kind
		}
	}
	if kind == 0 && tok.previous(excluded) != nil {
		if s := tok.previous(excluded); s != nil {
			kind = s.kind
		}
	}
	// indent the last comment line at the same level than scenario and background
	if tok.next([]gherkin.TokenType{gherkin.TokenTypeEmpty}) == nil {
		return indent
	}
	return paddings[kind]
}

func isDescriptionFeature(tok *token) bool {
	excluded := []gherkin.TokenType{gherkin.TokenTypeEmpty}
	if tok.previous(excluded) != nil {
		if t := tok.previous(excluded); t != nil && t.kind == gherkin.TokenTypeFeatureLine {
			return true
		}
	}
	return false
}

func trimLinesSpace(lines []string) []string {
	var content []string
	for _, line := range lines {
		content = append(content, strings.TrimSpace(line))
	}
	return content
}

func trimExtraTrailingSpace(lines []string) []string {
	var content []string
	for _, line := range lines {
		content = append(content, strings.TrimRight(line, " \t"))
	}
	return content
}

func indentStrings(padding int, lines []string) []string {
	var content []string
	for _, line := range lines {
		content = append(content, strings.Repeat(" ", padding)+line)
	}
	return content
}

func extractTokensText(tokens []*gherkin.Token) []string {
	var content []string
	for _, token := range tokens {
		content = append(content, token.Text)
	}
	return content
}

func extractTokensItemsText(tokens []*gherkin.Token) []string {
	var content []string
	for _, token := range tokens {
		var t []string
		for _, item := range token.Items {
			t = append(t, item.Text)
		}
		content = append(content, strings.Join(t, " "))
	}
	return content
}

func extractTokensKeywordAndText(tokens []*gherkin.Token) []string {
	var content []string
	for _, token := range tokens {
		content = append(content, fmt.Sprintf("%s%s", token.Keyword, token.Text))
	}
	return content
}

func extractKeywordAndTextSeparatedWithAColon(tokens []*gherkin.Token) []string {
	var content []string
	for _, token := range tokens {
		content = append(content, fmt.Sprintf("%s: %s", token.Keyword, token.Text))
	}
	return content
}

func extractKeyword(tokens []*gherkin.Token) []string {
	var content []string
	for _, t := range tokens {
		content = append(content, t.Keyword)
	}
	return content
}

func extractTableRowsAndComments(tokens []*gherkin.Token) []string {
	type tableElement struct {
		content []string
		kind    gherkin.TokenType
	}

	var rows [][]string
	var tableElements []tableElement
	for _, token := range tokens {
		element := tableElement{}

		if token.Type == gherkin.TokenTypeComment {
			element.kind = token.Type
			element.content = []string{token.Text}
		} else {
			var row []string
			for _, data := range token.Items {
				var text string

				source := []byte(data.Text)
				if ok := json.Valid(source); ok == true {
					var buffer bytes.Buffer
					_ = json.Compact(&buffer, source)
					text = buffer.String()
				} else {
					text = data.Text
				}

				// A remaining pipe means it was escaped before to not be messed with pipe column delimiter
				// so here we introduce the escaping sequence back
				text = strings.ReplaceAll(text, "\\", "\\\\")
				text = strings.ReplaceAll(text, "\n", "\\n")
				text = strings.ReplaceAll(text, "|", "\\|")
				row = append(row, text)
			}
			element.kind = token.Type
			element.content = row
			rows = append(rows, row)
		}
		tableElements = append(tableElements, element)
	}

	var tableRows []string
	lengths := calculateLonguestLineLengthPerColumn(rows)
	for _, tableElement := range tableElements {
		var inputs []interface{}
		fmtDirective := ""

		if tableElement.kind == gherkin.TokenTypeComment {
			inputs = append(inputs, trimLinesSpace(tableElement.content)[0])
			fmtDirective = "%s"
		} else {
			for i, str := range tableElement.content {
				inputs = append(inputs, str)
				fmtDirective += "| %-" + strconv.Itoa(lengths[i]) + "s "
			}
			fmtDirective += "|"
		}
		tableRows = append(tableRows, fmt.Sprintf(fmtDirective, inputs...))
	}
	return tableRows
}

func calculateLonguestLineLengthPerColumn(rows [][]string) []int {
	var lengths []int
	for i, row := range rows {
		for j, str := range row {
			switch true {
			case i == 0:
				lengths = append(lengths, utf8.RuneCountInString(str))
			case i != 0 && len(lengths) > j && lengths[j] < utf8.RuneCountInString(str):
				lengths[j] = utf8.RuneCountInString(str)
			default:
				lengths = append(lengths, 0)
			}
		}
	}
	return lengths
}
