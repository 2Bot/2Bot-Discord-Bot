package main

import (
	"fmt"
	l "log"
	"os"
)

type logging struct {
	i *l.Logger
	e *l.Logger
	t *l.Logger
}

func newLog() logging {
	return logging{
		l.New(os.Stdout, "INFO  - ", l.Ldate|l.Ltime),
		l.New(os.Stdout, "ERROR - ", l.Ldate|l.Ltime),
		l.New(os.Stdout, "TRACE - ", l.Ldate|l.Ltime),
	}
}

func (l logging) Error(f string, s ...interface{}) {
	l.e.Print(f, " ", fmt.Sprintln(s...))
}

func (l logging) Info(f string, s ...interface{}) {
	l.i.Print(f, " ", fmt.Sprintln(s...))
}

func (l logging) Trace(f string, s ...interface{}) {
	l.t.Print(f, " ", fmt.Sprintln(s...))
}
