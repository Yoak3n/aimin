package logger

import (
	"bytes"
	"dna/pkg/util"
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

type Formatter struct{}
type Hook struct {
	file       *os.File
	errorFile  *os.File
	driverFile *os.File
}

func (h *Hook) Levels() []logrus.Level { return logrus.AllLevels }
func (h *Hook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	if entry.Level == logrus.ErrorLevel || entry.Level == logrus.FatalLevel || entry.Level == logrus.PanicLevel {
		_, err = h.errorFile.Write([]byte(line))
	}
	//if strings.HasSuffix(path.Dir(entry.Caller.File), "blivedanmu") {
	//	h.driverFile.Write([]byte(line))
	//}
	_, err = h.file.Write([]byte(line))
	return err
}

func initHook(h *Hook) {
	var err error
	h.file, err = os.OpenFile("data/log/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	h.errorFile, err = os.OpenFile("data/log/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	//h.driverFile, err = os.OpenFile("data/log/runtime.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err != nil {
	//	panic(err)
	//}
}

func (Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	level := ""
	switch entry.Level {
	case logrus.InfoLevel:
		level = "INFO"
	case logrus.WarnLevel:
		level = "WARN"
	case logrus.ErrorLevel:
		level = "ERROR"
	case logrus.FatalLevel:
		level = "FATAL"
	case logrus.PanicLevel:
		level = "PANIC"
	default:
		level = "DEBUG"
	}
	var buf *bytes.Buffer
	if entry.Buffer == nil {
		buf = &bytes.Buffer{}
	} else {
		buf = entry.Buffer
	}
	dirPath := util.GetFileDir(path.Dir(entry.Caller.File))
	fileValue := fmt.Sprintf("%s/%s:%d", dirPath, path.Base(entry.Caller.File), entry.Caller.Line)
	fmt.Fprintf(buf, "[%s][%s]%s %s\n", level, entry.Time.Format("1-2|15:04:05"), fileValue, entry.Message)
	return buf.Bytes(), nil
}

func Init() {
	_ = util.CreateDirNotExists("data/log")
	Logger = logrus.New()
	hook := &Hook{}
	initHook(hook)
	Logger.AddHook(hook)
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&Formatter{})
}
