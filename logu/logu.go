// Package is a simple wrapper on top of default "log/slog". Provides handy logging utility functions. Not really that
// great for big projects, but better than ignoring errors or blindly propagating them.
package logu

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
)

// FullStackTrace, when set to true, enables the inclusion of a full stack trace in logged error messages.
var FullStackTrace = false

func getCallerLoc() (string, int) {
	var stack [10]uintptr
	runtime.Callers(3, stack[:])
	f := runtime.FuncForPC(stack[0])
	file, line := f.FileLine(stack[0])
	return file, line
}

func getStackTrace() []string {
	var stack [50]uintptr
	var out []string
	fileToSkip := ""
	for i, pc := range stack[:runtime.Callers(0, stack[:])] {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		if i == 0 {
			fileToSkip = file
		}
		if file == fileToSkip {
			continue
		}
		out = append(out, fmt.Sprintf("%s:%d", file, line))
	}
	return out
}

type Context[T any] struct {
	v        T
	err      error
	msg      string
	exitCode int
}

// Context constructor for `err := ...` scenario. Don't forget to call Do()!
func One(err error) Context[bool] {
	return Context[bool]{v: err != nil, err: err}
}

// Context constructor for `v, err := ...` scenario. Don't forget to call Do()!
func Two[T any](v T, err error) Context[T] {
	return Context[T]{v: v, err: err}
}

// Context constructor for `v, ok := ...` scenario. Don't forget to call Do()!
func Okay[T any](v T, ok bool) Context[T] {
	var err error
	if !ok {
		err = errors.New("invalid/missing value")
	}
	return Context[T]{v: v, err: err}
}

func (c Context[T]) getMessage() string {
	if c.msg != "" {
		return c.msg
	}
	if c.exitCode != 0 {
		return "fatal error"
	} else {
		return "error"
	}
}

// When Do() is called, the following message will be logged on error instead of the default.
func (c Context[T]) Message(msg string) Context[T] {
	c.msg = msg
	return c
}

// Check the error and log the message if error is not nil.
func (c Context[T]) Do() T {
	if c.err != nil {
		if FullStackTrace {
			slog.Error(c.getMessage(), "error", c.err, "stacktrace", getStackTrace())
		} else {
			slog.Error(c.getMessage(), "error", c.err)
		}
		if c.exitCode != 0 {
			os.Exit(c.exitCode)
		}
	}
	return c.v
}

// When Do() is called, the context will also call os.Exit(exitCode) at the end.
func (c Context[T]) Fatal(exitCode int) Context[T] {
	c.exitCode = exitCode
	return c
}

// Shortcut for: One(err).Message("<file>:<line>").Do().
func Err(err error) bool {
	file, line := getCallerLoc()
	return One(err).Message(fmt.Sprintf("%s:%d", file, line)).Do()
}

// Shortcut for: One(err).Message("<file>:<line>").Do(), but returns an error.
func Report(err error) error {
	file, line := getCallerLoc()
	One(err).Message(fmt.Sprintf("%s:%d", file, line)).Do()
	return err
}

// Shortcut for: Two(err).Message("<file>:<line>").Do().
func Err2[T any](v T, err error) T {
	file, line := getCallerLoc()
	return Two(v, err).Message(fmt.Sprintf("%s:%d", file, line)).Do()
}

// Shortcut for: Okay(v, ok).Message("<file>:<line>").Do().
func OK[T any](v T, ok bool) T {
	file, line := getCallerLoc()
	return Okay(v, ok).Message(fmt.Sprintf("%s:%d", file, line)).Do()
}

// Shortcut for: One(err).Fatal(1).Message("<file>:<line>").Do().
func FatalErr(err error) bool {
	file, line := getCallerLoc()
	return One(err).Fatal(1).Message(fmt.Sprintf("%s:%d", file, line)).Do()
}

// Shortcut for: Two(err).Fatal(1).Message("<file>:<line>").Do().
func FatalErr2[T any](v T, err error) T {
	file, line := getCallerLoc()
	return Two(v, err).Fatal(1).Message(fmt.Sprintf("%s:%d", file, line)).Do()
}

// Shortcut for: Okay(v, ok).Fatal(1).Message("<file>:<line>").Do().
func FatalOK[T any](v T, ok bool) T {
	file, line := getCallerLoc()
	return Okay(v, ok).Fatal(1).Message(fmt.Sprintf("%s:%d", file, line)).Do()
}
