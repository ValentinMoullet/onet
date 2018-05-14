// Package log is an output-library that can print nicely formatted
// messages to the screen.
//
// There are log-level messages that will be printed according to the
// current debug-level set. Furthermore a set of common messages exist
// that are printed according to a chosen format.
//
// The log-level messages are:
//	log.Lvl1("Important information")
//	log.Lvl2("Less important information")
//	log.Lvl3("Eventually flooding information")
//	log.Lvl4("Definitively flooding information")
//	log.Lvl5("I hope you never need this")
// in your program, then according to the debug-level one or more levels of
// output will be shown. To set the debug-level, use
//	log.SetDebugVisible(3)
// which will show all `Lvl1`, `Lvl2`, and `Lvl3`. If you want to turn
// on just one output, you can use
//	log.LLvl2("Less important information")
// By adding a single 'L' to the method, it *always* gets printed.
//
// You can also add a 'f' to the name and use it like fmt.Printf:
//	log.Lvlf1("Level: %d/%d", now, max)
//
// The common messages are:
//	log.Print("Simple output")
//	log.Info("For your information")
//	log.Warn("Only a warning")
//	log.Error("This is an error, but continues")
//	log.Panic("Something really went bad - calls panic")
//	log.Fatal("No way to continue - calls os.Exit")
//
// These messages are printed according to the value of 'Format':
// - Format == FormatLvl - same as log.Lvl
// - Format == FormatPython - with some nice python-style formatting
// - Format == FormatNone - just as plain text
//
// The log-package also takes into account the following environment-variables:
//	DEBUG_LVL // will act like SetDebugVisible
//	DEBUG_TIME // if 'true' it will print the date and time
//	DEBUG_COLOR // if 'false' it will not use colors
// But for this the function ParseEnv() or AddFlags() has to be called.
package log

import (
	"bytes"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"time"
)

// For testing we can change the output-writer
var stdOut io.Writer
var stdErr io.Writer

func init() {
	stdOut = os.Stdout
	stdErr = os.Stderr
}

var bufStdOut bytes.Buffer
var bufStdErr bytes.Buffer

// OutputToBuf is called for sending all the log.*-outputs to internal buffers
// that can be used for checking what the logger would've written. This is
// mostly used for tests. The buffers are zeroed after this call.
func OutputToBuf() {
	debugMut.Lock()
	defer debugMut.Unlock()
	stdOut = &bufStdOut
	stdErr = &bufStdErr
	bufStdOut.Reset()
	bufStdErr.Reset()
}

// OutputToOs redirects the output of the log.*-outputs again to the os.
func OutputToOs() {
	debugMut.Lock()
	defer debugMut.Unlock()
	stdOut = os.Stdout
	stdErr = os.Stderr
}

// GetStdOut returns all log.*-outputs to StdOut since the last call.
func GetStdOut() string {
	debugMut.Lock()
	defer debugMut.Unlock()
	ret := bufStdOut.String()
	bufStdOut.Reset()
	return ret
}

// GetStdErr returns all log.*-outputs to StdErr since the last call.
func GetStdErr() string {
	debugMut.Lock()
	defer debugMut.Unlock()
	ret := bufStdErr.String()
	bufStdErr.Reset()
	return ret
}


type fileLogger struct {
	path string
}


func (fl *fileLogger) Log(level int, msg string) {
	// The file should exist (created when calling 'NewFileLogger')
	f, err := os.OpenFile(fl.path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
	    panic(err)
	}

	defer f.Close()

	str := msg + "\n"
	if showTime {
		ti := time.Now()
		str = fmt.Sprintf("[%s.%09d] %s", ti.Format("06/02/01 15:04:05"), ti.Nanosecond(), str)
	}

	if _, err = f.WriteString(str); err != nil {
	    panic(err)
	}
}

func NewFileLogger(path string) error {
	// Override file if it already exists.
	_, err := os.Create(path)
	if err != nil {
		return err
	}
	fl := &fileLogger{path: path}
	RegisterListener(fl)
	return nil
}

type syslogLogger struct {
	writer *syslog.Writer
}

func (sl *syslogLogger) Log(level int, msg string) {
	_, err := sl.writer.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}

func NewSyslogLogger(priority syslog.Priority, tag string) (*syslog.Writer, error) {
	writer, err := syslog.New(priority, tag)
	if err != nil {
		return nil, err
	}
	sl := &syslogLogger{writer: writer}
	RegisterListener(sl)
	return writer, nil
}
