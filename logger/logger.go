// The oryx logger package provides connection-oriented log service.
//		logger.I(ctx, ...)
//		logger.T(ctx, ...)
//		logger.W(ctx, ...)
//		logger.E(ctx, ...)
// Or use format:
//		logger.If(ctx, format, ...)
//		logger.Tf(ctx, format, ...)
//		logger.Wf(ctx, format, ...)
//		logger.Ef(ctx, format, ...)
// @remark the Context is optional thus can be nil.
// @remark From 1.7+, the ctx could be context.Context, wrap by logger.WithContext,
// 	please read ExampleLogger_ContextGO17().
package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// default level for logger.
const (
	logInfoLabel  = "[info] "
	logTraceLabel = "[trace] "
	logWarnLabel  = "[warn] "
	logErrorLabel = "[error] "
)

// The context for current goroutine.
// It maybe a cidContext or context.Context from GO1.7.
// @remark Use logger.WithContext(ctx) to wrap the context.
type Context interface{}

// The context to get current coroutine cid.
type cidContext interface {
	Cid() int
}

// the LOG+ which provides connection-based log.
type loggerPlus struct {
	logger *log.Logger
}

func NewLoggerPlus(l *log.Logger) Logger {
	return &loggerPlus{logger: l}
}

func (v *loggerPlus) format(ctx Context, a ...interface{}) []interface{} {
	if ctx == nil {
		return append([]interface{}{fmt.Sprintf("[%v] ", os.Getpid())}, a...)
	} else if ctx, ok := ctx.(cidContext); ok {
		return append([]interface{}{fmt.Sprintf("[%v][%v] ", os.Getpid(), ctx.Cid())}, a...)
	}
	return a
}

func (v *loggerPlus) formatf(ctx Context, format string, a ...interface{}) (string, []interface{}) {
	if ctx == nil {
		return "[%v] " + format, append([]interface{}{os.Getpid()}, a...)
	} else if ctx, ok := ctx.(cidContext); ok {
		return "[%v][%v] " + format, append([]interface{}{os.Getpid(), ctx.Cid()}, a...)
	}
	return format, a
}

var colorYellow = "\033[33m"
var colorRed = "\033[31m"
var colorBlack = "\033[0m"

func (v *loggerPlus) doPrintln(args ...interface{}) {
	if previousIo == nil {
		if v == Error {
			fmt.Fprintf(os.Stdout, colorRed)
			v.logger.Println(args...)
			fmt.Fprintf(os.Stdout, colorBlack)
		} else if v == Warn {
			fmt.Fprintf(os.Stdout, colorYellow)
			v.logger.Println(args...)
			fmt.Fprintf(os.Stdout, colorBlack)
		} else {
			v.logger.Println(args...)
		}
	} else {
		v.logger.Println(args...)
	}
}

func (v *loggerPlus) doPrintf(format string, args ...interface{}) {
	if previousIo == nil {
		if v == Error {
			fmt.Fprintf(os.Stdout, colorRed)
			v.logger.Printf(format, args...)
			fmt.Fprintf(os.Stdout, colorBlack)
		} else if v == Warn {
			fmt.Fprintf(os.Stdout, colorYellow)
			v.logger.Printf(format, args...)
			fmt.Fprintf(os.Stdout, colorBlack)
		} else {
			v.logger.Printf(format, args...)
		}
	} else {
		v.logger.Printf(format, args...)
	}
}

// Info, the verbose info level, very detail log, the lowest level, to discard.
var Info Logger

// Alias for Info level println.
func I(ctx Context, a ...interface{}) {
	Info.Println(ctx, a...)
}

// Printf for Info level log.
func If(ctx Context, format string, a ...interface{}) {
	Info.Printf(ctx, format, a...)
}

// Trace, the trace level, something important, the default log level, to stdout.
var Trace Logger

// Alias for Trace level println.
func T(ctx Context, a ...interface{}) {
	Trace.Println(ctx, a...)
}

// Printf for Trace level log.
func Tf(ctx Context, format string, a ...interface{}) {
	Trace.Printf(ctx, format, a...)
}

// Warn, the warning level, dangerous information, to Stdout.
var Warn Logger

// Alias for Warn level println.
func W(ctx Context, a ...interface{}) {
	Warn.Println(ctx, a...)
}

// Printf for Warn level log.
func Wf(ctx Context, format string, a ...interface{}) {
	Warn.Printf(ctx, format, a...)
}

// Error, the error level, fatal error things, ot Stdout.
var Error Logger

// Alias for Error level println.
func E(ctx Context, a ...interface{}) {
	Error.Println(ctx, a...)
}

// Printf for Error level log.
func Ef(ctx Context, format string, a ...interface{}) {
	Error.Printf(ctx, format, a...)
}

// The logger for oryx.
type Logger interface {
	// Println for logger plus,
	// @param ctx the connection-oriented context,
	// 	or context.Context from GO1.7, or nil to ignore.
	Println(ctx Context, a ...interface{})
	Printf(ctx Context, format string, a ...interface{})
}

func init() {
	Info = NewLoggerPlus(log.New(ioutil.Discard, logInfoLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Trace = NewLoggerPlus(log.New(os.Stdout, logTraceLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Warn = NewLoggerPlus(log.New(os.Stdout, logWarnLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Error = NewLoggerPlus(log.New(os.Stdout, logErrorLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
}

// Switch the underlayer io.
// @remark user must close previous io for logger never close it.
func Switch(w io.Writer) {
	// TODO: support level, default to trace here.
	Info = NewLoggerPlus(log.New(ioutil.Discard, logInfoLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Trace = NewLoggerPlus(log.New(w, logTraceLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Warn = NewLoggerPlus(log.New(w, logWarnLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Error = NewLoggerPlus(log.New(w, logErrorLabel, log.Ldate|log.Ltime|log.Lmicroseconds))

	if w, ok := w.(io.Closer); ok {
		previousIo = w
	}
}

// The previous underlayer io for logger.
var previousIo io.Closer

// The interface io.Closer
// Cleanup the logger, discard any log util switch to fresh writer.
func Close() (err error) {
	Info = NewLoggerPlus(log.New(ioutil.Discard, logInfoLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Trace = NewLoggerPlus(log.New(ioutil.Discard, logTraceLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Warn = NewLoggerPlus(log.New(ioutil.Discard, logWarnLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Error = NewLoggerPlus(log.New(ioutil.Discard, logErrorLabel, log.Ldate|log.Ltime|log.Lmicroseconds))

	if previousIo != nil {
		err = previousIo.Close()
		previousIo = nil
	}

	return
}
