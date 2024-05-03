package formatter

import (
	"log"
	"os"

	"github.com/fatih/color"
)

type Log interface {
	Print(str string)
	Error(err error)
	Success(str string)
}

var logger = log.New(os.Stderr, "", log.Lmsgprefix)

type Logger struct{}

func NewLogger() Log {
	return Logger{}
}

func (l Logger) Print(str string) {
	logger.Println(color.New(color.FgWhite).Sprint(str))
}

func (l Logger) Error(err error) {
	logger.Println(color.New(color.FgRed).Sprint(err.Error()))
}

func (l Logger) Success(str string) {
	logger.Println(color.New(color.FgGreen).Sprint(str))
}
