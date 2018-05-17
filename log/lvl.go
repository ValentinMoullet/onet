package log

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/daviddengcn/go-colortext"
)

const (
	lvlWarning = iota - 20
	lvlError
	lvlFatal
	lvlPanic
	lvlInfo
	lvlPrint
)

// These formats can be used in place of the debugVisible
const (
	// FormatPython uses [x] and others to indicate what is shown
	FormatPython = -1
	// FormatNone is just pure print
	FormatNone = 0
)

// defaultMainTest indicates what debug-level should be used when `go test -v`
// is called.
const defaultMainTest = 2

// MainTestWait is the maximum time the MainTest-method waits before aborting
// and printing a stack-trace of all functions.
// This is deprecated and should not be used anymore.
var MainTestWait = 0 * time.Minute

// NamePadding - the padding of functions to make a nice debug-output - this is automatically updated
// whenever there are longer functions and kept at that new maximum. If you prefer
// to have a fixed output and don't remember oversized names, put a negative value
// in here.
var NamePadding = 40

// LinePadding of line-numbers for a nice debug-output - used in the same way as
// NamePadding.
var LinePadding = 3

// StaticMsg - if this variable is set, it will be outputted between the
// position and the message.
var StaticMsg = ""

// outputLines can be false to suppress outputting of lines in tests.
var outputLines = true

var debugMut sync.RWMutex

var regexpPaths, _ = regexp.Compile(".*/")

func init() {
	newStdLogger()
	ParseEnv()
}

func lvl(lvl, skip int, args ...interface{}) {
	debugMut.Lock()
	defer debugMut.Unlock()
	for _, l := range loggers {
		if lvl > l.debugLvl {
			continue
		}

		pc, _, line, _ := runtime.Caller(skip)
		name := regexpPaths.ReplaceAllString(runtime.FuncForPC(pc).Name(), "")
		lineStr := fmt.Sprintf("%d", line)

		// For the testing-framework, we check the resulting string. So as not to
		// have the tests fail every time somebody moves the functions, we put
		// the line-# to 0
		if !outputLines {
			line = 0
		}

		fmtstr := ""
		if l.useColors {
			// Only adjust the name and line padding if we also have color.
			if len(name) > NamePadding && NamePadding > 0 {
				NamePadding = len(name)
			}
			if len(lineStr) > LinePadding && LinePadding > 0 {
				LinePadding = len(name)
			}
		}
		fmtstr = fmt.Sprintf("%%%ds: %%%dd", NamePadding, LinePadding)
		caller := fmt.Sprintf(fmtstr, name, line)
		if StaticMsg != "" {
			caller += "@" + StaticMsg
		}
		message := fmt.Sprintln(args...)
		bright := lvl < 0
		lvlAbs := lvl
		if bright {
			lvlAbs *= -1
		}

		lvlStr := strconv.Itoa(lvlAbs)
		if lvl < 0 {
			lvlStr += "!"
		}
		switch lvl {
		case lvlPrint:
			fg(l, ct.White, true)
			lvlStr = "I"
		case lvlInfo:
			fg(l, ct.White, true)
			lvlStr = "I"
		case lvlWarning:
			fg(l, ct.Green, true)
			lvlStr = "W"
		case lvlError:
			fg(l, ct.Red, false)
			lvlStr = "E"
		case lvlFatal:
			fg(l, ct.Red, true)
			lvlStr = "F"
		case lvlPanic:
			fg(l, ct.Red, true)
			lvlStr = "P"
		default:
			if lvl != 0 {
				if lvlAbs <= 5 {
					colors := []ct.Color{ct.Yellow, ct.Cyan, ct.Green, ct.Blue, ct.Cyan}
					fg(l, colors[lvlAbs-1], bright)
				}
			}
		}
		str := fmt.Sprintf(": (%s) - %s", caller, message)
		if l.showTime {
			ti := time.Now()
			str = fmt.Sprintf("%s.%09d%s", ti.Format("06/02/01 15:04:05"), ti.Nanosecond(), str)
		}
		str = fmt.Sprintf("%-2s%s", lvlStr, str)

		l.Log(lvl, str)

		if l.useColors {
			ct.ResetColor()
		}
	}
}

func fg(l *LoggerInfo, c ct.Color, bright bool) {
	if l.useColors {
		ct.Foreground(c, bright)
	}
}

// Needs two functions to keep the caller-depth the same and find who calls us
// Lvlf1 -> Lvlf -> lvl
// or
// Lvl1 -> lvld -> lvl
func lvlf(l int, f string, args ...interface{}) {
	if l > DebugVisible() {
		return
	}
	lvl(l, 3, fmt.Sprintf(f, args...))
}
func lvld(l int, args ...interface{}) {
	lvl(l, 3, args...)
}

// Lvl1 debug output is informational and always displayed
func Lvl1(args ...interface{}) {
	lvld(1, args...)
}

// Lvl2 is more verbose but doesn't spam the stdout in case
// there is a big simulation
func Lvl2(args ...interface{}) {
	lvld(2, args...)
}

// Lvl3 gives debug-output that can make it difficult to read
// for big simulations with more than 100 hosts
func Lvl3(args ...interface{}) {
	lvld(3, args...)
}

// Lvl4 is only good for test-runs with very limited output
func Lvl4(args ...interface{}) {
	lvld(4, args...)
}

// Lvl5 is for big data
func Lvl5(args ...interface{}) {
	lvld(5, args...)
}

// Lvlf1 is like Lvl1 but with a format-string
func Lvlf1(f string, args ...interface{}) {
	lvlf(1, f, args...)
}

// Lvlf2 is like Lvl2 but with a format-string
func Lvlf2(f string, args ...interface{}) {
	lvlf(2, f, args...)
}

// Lvlf3 is like Lvl3 but with a format-string
func Lvlf3(f string, args ...interface{}) {
	lvlf(3, f, args...)
}

// Lvlf4 is like Lvl4 but with a format-string
func Lvlf4(f string, args ...interface{}) {
	lvlf(4, f, args...)
}

// Lvlf5 is like Lvl5 but with a format-string
func Lvlf5(f string, args ...interface{}) {
	lvlf(5, f, args...)
}

// LLvl1 *always* prints
func LLvl1(args ...interface{}) { lvld(-1, args...) }

// LLvl2 *always* prints
func LLvl2(args ...interface{}) { lvld(-2, args...) }

// LLvl3 *always* prints
func LLvl3(args ...interface{}) { lvld(-3, args...) }

// LLvl4 *always* prints
func LLvl4(args ...interface{}) { lvld(-4, args...) }

// LLvl5 *always* prints
func LLvl5(args ...interface{}) { lvld(-5, args...) }

// LLvlf1 *always* prints
func LLvlf1(f string, args ...interface{}) { lvlf(-1, f, args...) }

// LLvlf2 *always* prints
func LLvlf2(f string, args ...interface{}) { lvlf(-2, f, args...) }

// LLvlf3 *always* prints
func LLvlf3(f string, args ...interface{}) { lvlf(-3, f, args...) }

// LLvlf4 *always* prints
func LLvlf4(f string, args ...interface{}) { lvlf(-4, f, args...) }

// LLvlf5 *always* prints
func LLvlf5(f string, args ...interface{}) { lvlf(-5, f, args...) }

// TestOutput sets the DebugVisible to 0 if 'show'
// is false, else it will set DebugVisible to 'level'
//
// Usage: TestOutput( test.Verbose(), 2 )
func TestOutput(show bool, level int) {
	debugMut.Lock()
	defer debugMut.Unlock()

	if show {
		loggers[0].debugLvl = level
	} else {
		loggers[0].debugLvl = 0
	}
}

// SetDebugVisible set the global debug output level in a go-rountine-safe way
func SetDebugVisible(lvl int) {
	debugMut.Lock()
	defer debugMut.Unlock()
	loggers[0].debugLvl = lvl
}

// DebugVisible returns the actual visible debug-level
func DebugVisible() int {
	debugMut.RLock()
	defer debugMut.RUnlock()
	return loggers[0].debugLvl
}

// SetShowTime allows for turning on the flag that adds the current
// time to the debug-output
func SetShowTime(show bool) {
	debugMut.Lock()
	defer debugMut.Unlock()
	loggers[0].showTime = show
}

// ShowTime returns the current setting for showing the time in the debug
// output
func ShowTime() bool {
	debugMut.Lock()
	defer debugMut.Unlock()
	return loggers[0].showTime
}

// SetUseColors can turn off or turn on the use of colors in the debug-output
func SetUseColors(show bool) {
	debugMut.Lock()
	defer debugMut.Unlock()
	loggers[0].useColors = show
}

// UseColors returns the actual setting of the color-usage in log
func UseColors() bool {
	debugMut.Lock()
	defer debugMut.Unlock()
	return loggers[0].useColors
}

// MainTest can be called from TestMain. It will parse the flags and
// set the DebugVisible to defaultMainTest, then run the tests and check for
// remaining go-routines.
// If you give it an integer as optional parameter, it will set the
// debig-lvl to that integer. As `go test` will only output whenever
// `-v` is given, this gives no disadvantage over setting the default
// output-level.
func MainTest(m *testing.M, ls ...int) {
	flag.Parse()
	l := defaultMainTest
	if len(ls) > 0 {
		l = ls[0]
	}
	TestOutput(testing.Verbose(), l)
	done := make(chan int)
	go func() {
		code := m.Run()
		done <- code
	}()
	select {
	case code := <-done:
		AfterTest(nil)
		os.Exit(code)
	case <-time.After(interpretWait()):
		Error("Didn't finish in time")
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		os.Exit(1)
	}
}

// ParseEnv looks at the following environment-variables:
//   DEBUG_LVL - for the actual debug-lvl - default is 1
//   DEBUG_TIME - whether to show the timestamp - default is false
//   DEBUG_COLOR - whether to color the output - default is false
func ParseEnv() {
	dv := os.Getenv("DEBUG_LVL")
	if dv != "" {
		dvInt, err := strconv.Atoi(dv)
		Lvl3("Setting level to", dv, dvInt, err)
		SetDebugVisible(dvInt)
		if err != nil {
			Error("Couldn't convert", dv, "to debug-level")
		}
	}
	dt := os.Getenv("DEBUG_TIME")
	if dt != "" {
		dtInt, err := strconv.ParseBool(dt)
		Lvl3("Setting showTime to", dt, dtInt, err)
		SetShowTime(dtInt)
		if err != nil {
			Error("Couldn't convert", dt, "to boolean")
		}
	}
	dc := os.Getenv("DEBUG_COLOR")
	if dc != "" {
		ucInt, err := strconv.ParseBool(dc)
		Lvl3("Setting useColor to", dc, ucInt, err)
		SetUseColors(ucInt)
		if err != nil {
			Error("Couldn't convert", dc, "to boolean")
		}
	}
}

// RegisterFlags adds the flags and the variables for the debug-control
// (standard logger) using the standard flag-package.
func RegisterFlags() {
	ParseEnv()
	defaultDebugLvl := DebugVisible()
	defaultShowTime := ShowTime()
	defaultUseColors := UseColors()
	debugMut.Lock()
	defer debugMut.Unlock()
	flag.IntVar(&loggers[0].debugLvl, "debug", defaultDebugLvl, "Change debug level (0-5)")
	flag.BoolVar(&loggers[0].showTime, "debug-time", defaultShowTime, "Shows the time of each message")
	flag.BoolVar(&loggers[0].useColors, "debug-color", defaultUseColors, "Colors each message")
}

var timeoutFlagMutex sync.Mutex

// interpretWait will use the test.timeout flag and MainTestWait to get
// the time to wait. From highest preference to lowest preference:
//  1. "-timeout 1m"
//  2. MainTestWait = 2*time.Minute
//  3. 10*time.Minute
// interpretWait will throw a warning if MainTestWait has been set.
func interpretWait() time.Duration {
	timeoutFlagMutex.Lock()
	defer timeoutFlagMutex.Unlock()
	toFlag := flag.Lookup("test.timeout")
	if toFlag == nil {
		Fatal("MainTest should not be called outside of tests")
	}
	dur := MainTestWait
	var err error
	if dur != 0*time.Second {
		Warn("Usage of MainTestWait is deprecated!")
	} else {
		dur = 10 * time.Minute
	}
	if toStr := toFlag.Value.String(); toStr != "0s" {
		dur, err = time.ParseDuration(toStr)
		ErrFatal(err, "couldn't parse passed timeout value")
	}
	return dur
}
