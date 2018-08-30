package logging

import (
	"os"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("2Bot")

func init() {
	errorLog := logging.NewLogBackend(os.Stderr, "ERROR", 0)
	infoLog := logging.NewLogBackend(os.Stdout, "INFO", 0)
	traceLog := logging.NewLogBackend(os.Stdout, "TRACE", 0)

}
