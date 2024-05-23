package log

import (
	baselog "log"
	"os"

	"github.com/fatih/color"
)

type logging struct{}

var logger = baselog.New(os.Stderr, "", baselog.Lmsgprefix)
var log logging

func Error(err error) {
	log.error(err)
}

func Success(str string) {
	log.success(str)
}

// GetLogger using for test purposes only. Do not use it for production code
func GetLogger() *baselog.Logger {
	return logger
}

func (l logging) error(err error) {
	logger.Println(color.New(color.FgRed).Sprint(err.Error()))
}

func (l logging) success(str string) {
	logger.Println(color.New(color.FgGreen).Sprint(str))
}
