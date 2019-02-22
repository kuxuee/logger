package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Level int32

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

const (
	INFINITE int = 0
)

/*
===================
 utils functions
===================
*/
func fileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

/*
===================
 log handlers
===================
*/
type Handler interface {
	SetOutput(w io.Writer)
	Output(calldepth int, s string) error
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})

	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})

	Flags() int
	SetFlags(flag int)
	Prefix() string
	SetPrefix(prefix string)
	close()
}

type LogHandler struct {
	lg *log.Logger
	mu sync.Mutex
}

type ConsoleHander struct {
	LogHandler
}

type FileHandler struct {
	LogHandler
	logfile *os.File
}

type RotatingHandler struct {
	LogHandler
	dir      string
	filename string
	maxNum   int
	maxSize  int64
	filetime string
	suffix   int
	logfile  *os.File
}

var Console, _ = NewConsoleHandler()

func NewConsoleHandler() (*ConsoleHander, error) {
	l := log.New(os.Stderr, "", log.LstdFlags)
	return &ConsoleHander{LogHandler: LogHandler{lg: l}}, nil
}

func NewFileHandler(filepath string) (*FileHandler, error) {
	logfile, _ := os.OpenFile(filepath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	l := log.New(logfile, "", log.LstdFlags)
	return &FileHandler{
		LogHandler: LogHandler{lg: l},
		logfile:    logfile,
	}, nil
}

func NewRotatingHandler(dir string, filename string, maxNum int, maxSize int64) (*RotatingHandler, error) {
	if maxNum < 0 {
		return nil, errors.Errorf("maxNum is less than 0")
	}

	h := &RotatingHandler{
		dir:      dir,
		filename: filename,
		maxNum:   maxNum,
		maxSize:  maxSize,
		suffix:   0,
	}
	h.newFileTime()

	logfile, _ := os.OpenFile(h.generateFileName(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	l := log.New(logfile, "", log.LstdFlags)
	h.LogHandler = LogHandler{lg: l}
	h.logfile = logfile

	if h.isMustRename() {
		h.rename()
	}

	// monitor filesize per second
	go func() {
		timer := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timer.C:
				h.fileCheck()
			}
		}
	}()

	return h, nil
}

func (l *LogHandler) SetOutput(w io.Writer) {
	l.lg.SetOutput(w)
}

func (l *LogHandler) Output(calldepth int, s string) error {
	return l.lg.Output(calldepth, s)
}

func (l *LogHandler) Printf(format string, v ...interface{}) {
	l.lg.Printf(format, v...)
}

func (l *LogHandler) Print(v ...interface{}) { l.lg.Print(v...) }

func (l *LogHandler) Println(v ...interface{}) { l.lg.Println(v...) }

func (l *LogHandler) Fatal(v ...interface{}) {
	l.lg.Output(3, fmt.Sprint(v...))
}

func (l *LogHandler) Fatalf(format string, v ...interface{}) {
	l.lg.Output(3, fmt.Sprintf(format, v...))
}

func (l *LogHandler) Fatalln(v ...interface{}) {
	l.lg.Output(3, fmt.Sprintln(v...))
}

func (l *LogHandler) Flags() int {
	return l.lg.Flags()
}

func (l *LogHandler) SetFlags(flag int) {
	l.lg.SetFlags(flag)
}

func (l *LogHandler) Prefix() string {
	return l.lg.Prefix()
}

func (l *LogHandler) SetPrefix(prefix string) {
	l.lg.SetPrefix(prefix)
}

func (l *LogHandler) Debug(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lg.Output(3, fmt.Sprintln("debug", v))
}

func (l *LogHandler) Info(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lg.Output(3, fmt.Sprintln("info", v))
}

func (l *LogHandler) Warn(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lg.Output(3, fmt.Sprintln("warn", v))
}

func (l *LogHandler) Error(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lg.Output(3, fmt.Sprintln("error", v))
}

func (l *LogHandler) close() {

}

func (h *FileHandler) close() {
	if h.logfile != nil {
		h.logfile.Close()
	}
}

func (h *RotatingHandler) close() {
	if h.logfile != nil {
		h.logfile.Close()
	}
}

func (h *RotatingHandler) isMustRename() bool {
	if fileSize(h.generateFileName()) >= h.maxSize {
		return true
	}
	return false
}

func (h *RotatingHandler) newFileTime() {
	t := time.Now()
	h.filetime = t.Format("20060102150405")
}

func (h *RotatingHandler) generateFileName() string {
	if h.suffix <= 0 {
		return fmt.Sprintf("%s/%s.%s.0.log", h.dir, h.filename, h.filetime)
	}
	return fmt.Sprintf("%s/%s.%s.%d.log", h.dir, h.filename, h.filetime, h.suffix)
}

func (h *RotatingHandler) rename() {
	if INFINITE == h.maxNum {
		h.suffix = h.suffix + 1
	} else {
		if 0 == (h.suffix+1)%h.maxNum {
			h.suffix = 0
		} else {
			h.suffix = h.suffix%h.maxNum + 1
		}
	}

	newpath := h.generateFileName()
	if isExist(newpath) {
		os.Remove(newpath)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.logfile != nil {
		h.logfile.Close()
	}

	h.logfile, _ = os.OpenFile(newpath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	h.lg.SetOutput(h.logfile)
}

func (h *RotatingHandler) fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if h.isMustRename() {
		h.rename()
	}
}

/*
===================
 logger
===================
*/
type _Logger struct {
	handlers []Handler
	level    Level
	mu       sync.Mutex
}

var logger = &_Logger{
	handlers: []Handler{
		Console,
	},
	level: DEBUG,
}

func SetHandlers(handlers ...Handler) {
	logger.handlers = handlers
}

func SetFlags(flag int) {
	for i := range logger.handlers {
		logger.handlers[i].SetFlags(flag)
	}
}

func SetLevel(level Level) {
	logger.level = level
}

func Print(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Print(v...)
	}
}

func Printf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Printf(format, v...)
	}
}

func Println(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Println(v...)
	}
}

func Fatal(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Fatal(v...)
	}
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Fatalf(format, v...)
	}
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Fatalln(v...)
	}
	os.Exit(1)
}

func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	for i := range logger.handlers {
		logger.handlers[i].Output(3, s)
	}
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	for i := range logger.handlers {
		logger.handlers[i].Output(3, s)
	}
	panic(s)
}

func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	for i := range logger.handlers {
		logger.handlers[i].Output(3, s)
	}
	panic(s)
}

func Debug(v ...interface{}) {
	if logger.level <= DEBUG {
		for i := range logger.handlers {
			logger.handlers[i].Debug(v...)
		}
	}
}

func Info(v ...interface{}) {
	if logger.level <= INFO {
		for i := range logger.handlers {
			logger.handlers[i].Info(v...)
		}
	}
}

func Warn(v ...interface{}) {
	if logger.level <= WARN {
		for i := range logger.handlers {
			logger.handlers[i].Warn(v...)
		}
	}
}

func Error(v ...interface{}) {
	if logger.level <= ERROR {
		for i := range logger.handlers {
			logger.handlers[i].Error(v...)
		}
	}
}

func Close() {
	for i := range logger.handlers {
		logger.handlers[i].close()
	}
}
