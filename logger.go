package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Level int

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
	filetime time.Time
	suffix   int
	logfile  *os.File
}

type logconfig struct {
	Handle   string `json:"handle"`
	Dir      string `json:"dir"`
	Filename string `json:"filename"`
	Maxnum   int    `json:"maxnum"`
	Maxsize  string `json:"maxsize"`
}

type logconfigs struct {
	Name  string      `json:"name"`
	Level int         `json:"level"`
	Data  []logconfig `json:"data"`
}

type configs struct {
	Logs []logconfigs `json:"logs"`
}

var Console, _ = NewConsoleHandler()

func NewConsoleHandler() (*ConsoleHander, error) {
	l := log.New(os.Stderr, "", log.LstdFlags)
	return &ConsoleHander{LogHandler: LogHandler{lg: l}}, nil
}

func NewFileHandler(filepath string) (*FileHandler, error) {
	i := strings.LastIndex(filepath, "\\")
	if -1 == i {
		i = strings.LastIndex(filepath, "/")
		if -1 == i {
			return nil, fmt.Errorf("Error filepath:%v", filepath)
		}
	}
	dir := filepath[:i]
	err := os.MkdirAll(dir, 0711)
	if err != nil {
		return nil, err
	}

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
	err := os.MkdirAll(dir, 0711)
	if err != nil {
		return nil, err
	}
	h := &RotatingHandler{
		dir:      dir,
		filename: filename,
		maxNum:   maxNum,
		maxSize:  maxSize,
		suffix:   0,
	}
	h.newFileData()

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

func newHandler(lg logconfig) (Handler, error) {
	if "console" == lg.Handle {
		return NewConsoleHandler()
	} else if "file" == lg.Handle {
		return NewFileHandler(lg.Filename)
	} else if "rotating" == lg.Handle {
		l := len(lg.Maxsize)
		if l < 3 {
			return nil, fmt.Errorf("Error maxsize:%v", lg.Maxsize)
		}
		unitStr := lg.Maxsize[l-2:]
		maxSizeStr := lg.Maxsize[:l-2]
		maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(unitStr) {
		case "mb":
			maxSize = maxSize * 1024 * 1024
		case "kb":
			maxSize = maxSize * 1024
		default:
			return nil, fmt.Errorf("Error maxsize type:%v", unitStr)
		}
		fmt.Println("maxsize:", maxSize)
		return NewRotatingHandler(lg.Dir, lg.Filename, lg.Maxnum, maxSize)
	}
	return nil, fmt.Errorf("Unknown handle:%v", lg.Handle)
}

func NewLogger(name string) error {
	filename := "./logs.config"
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	c := configs{}
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return err
	}

	level := 0
	for _, lgs := range c.Logs {
		if name == lgs.Name {
			level = lgs.Level
			if level < 0 || level > int(ERROR) {
				return fmt.Errorf("Unknown log name:%v level:%v", name, lgs.Level)
			}
			for _, lg := range lgs.Data {
				handler, err := newHandler(lg)
				if err != nil {
					Close()
					return err
				}
				logger.handlers = append(logger.handlers, handler)
			}
		}
	}
	if len(logger.handlers) <= 0 {
		return fmt.Errorf("Create logger error:%v", name)
	}

	SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
	SetLevel(Level(level))
	return nil
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
	t := time.Now()
	if t.Year() > h.filetime.Year() ||
		t.Year() == h.filetime.Year() && t.Month() > h.filetime.Month() ||
		t.Year() == h.filetime.Year() && t.Month() == h.filetime.Month() && t.Day() > h.filetime.Day() {

		h.newFileData()
		return true
	}
	return false
}

func (h *RotatingHandler) newFileData() {
	h.filetime = time.Now()
	if INFINITE == h.maxNum {
		h.suffix = -1
	} else {
		h.suffix = h.maxNum - 1
	}
}

func (h *RotatingHandler) generateFileName() string {
	filetime := h.filetime.Format("20060102150405")
	if h.suffix <= 0 {
		return fmt.Sprintf("%s/%s.%s.0.log", h.dir, h.filename, filetime)
	}
	return fmt.Sprintf("%s/%s.%s.%d.log", h.dir, h.filename, filetime, h.suffix)
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
	handlers: []Handler{},
	level:    DEBUG,
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
	logger.handlers = logger.handlers[0:0]
}
