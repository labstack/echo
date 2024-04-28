// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"io"

	"github.com/labstack/gommon/log"
)

// Logger defines the logging interface.
type Logger interface {
	Output() io.Writer
	SetOutput(w io.Writer)
	Prefix() string
	SetPrefix(p string)
	Level() log.Lvl
	SetLevel(v log.Lvl)
	SetHeader(h string)
	Print(i ...any)
	Printf(format string, args ...any)
	Printj(j log.JSON)
	Debug(i ...any)
	Debugf(format string, args ...any)
	Debugj(j log.JSON)
	Info(i ...any)
	Infof(format string, args ...any)
	Infoj(j log.JSON)
	Warn(i ...any)
	Warnf(format string, args ...any)
	Warnj(j log.JSON)
	Error(i ...any)
	Errorf(format string, args ...any)
	Errorj(j log.JSON)
	Fatal(i ...any)
	Fatalj(j log.JSON)
	Fatalf(format string, args ...any)
	Panic(i ...any)
	Panicj(j log.JSON)
	Panicf(format string, args ...any)
}
