package monitor

import (
	"io"
	"log"

	"github.com/gookit/color"
)

var (
	red    = color.FgLightRed.Render
	yellow = color.FgLightYellow.Render
)

type Monitor struct {
	Logs   bool
	Errors []error
	logger *log.Logger
}

func (m *Monitor) Fatal(err error) {
	if m.Logs {
		m.logger.Fatal(err)
	} else {
		m.logger.Fatalf("\n%s: %s\n\n", red("ERROR"), err)
	}
}

func (m *Monitor) Error(err error) {
	m.Errors = append(m.Errors, err)

	if m.Logs {
		m.logger.Print(err)
	} else {
		m.logger.Printf("%s: %s", red("ERROR"), err)
	}
}

func (m *Monitor) FileError(err error) {
	m.Errors = append(m.Errors, err)

	if m.Logs {
		m.logger.Print(err)
	} else {
		m.logger.Printf("^ %s: %s", red("ERROR"), err)
	}
}

func (m *Monitor) FileWarn(msg string) {
	if m.Logs {
		m.logger.Print(msg)
	} else {
		m.logger.Printf("^ %s: %s", yellow("WARN"), msg)
	}
}

func New(w io.Writer, logs bool) Monitor {
	l := log.New(w, "", 0)

	return Monitor{
		Logs:   logs,
		Errors: []error{},
		logger: l,
	}
}
