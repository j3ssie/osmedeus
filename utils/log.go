package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/kyokomi/emoji"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var logger = logrus.New()

// InitLog init log
func InitLog(options *libs.Options) {
	mwr := io.MultiWriter(os.Stdout)
	logDir := libs.LDIR
	if options.LogFile == "" {
		if !FolderExists(logDir) {
			os.MkdirAll(logDir, 0777)
		}
		tmpFile, err := os.CreateTemp(logDir, "osmedeus-*.log")
		if err == nil {
			options.LogFile = tmpFile.Name()
		} else {
			tmpFile, _ := os.CreateTemp("/tmp/", "osmedeus-*.log")
			options.LogFile = tmpFile.Name()
		}
	}

	logDir = filepath.Dir(options.LogFile)
	if !FolderExists(logDir) {
		os.MkdirAll(logDir, 0777)
	}

	f, err := os.OpenFile(options.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening log file: %v\n", options.LogFile)
		fmt.Fprintf(os.Stderr, "💡 You might want to switch to %v first via %v command", color.HiMagentaString("root user"), color.HiCyanString("sudo su"))
	} else {
		mwr = io.MultiWriter(os.Stdout, f)
	}

	logger = &logrus.Logger{
		Out:   mwr,
		Level: logrus.InfoLevel,
		Formatter: &prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05",
		},
	}

	if options.Debug == true {
		logger.SetLevel(logrus.DebugLevel)
		options.Verbose = true
	} else if options.Verbose == true {
		logger.SetLevel(logrus.ErrorLevel)
	} else if options.Quite == true {
		logger.SetOutput(io.Discard)
	}
}

// PrintLine print seperate line
func PrintLine() {
	dash := color.HiWhiteString("-")
	fmt.Println(strings.Repeat(dash, 40))
}

// GoodF print good message
func GoodF(format string, args ...interface{}) {
	good := color.HiGreenString("[+]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

// BannerF print info message
func BannerF(format string, data string) {
	banner := fmt.Sprintf("%v%v%v ", color.WhiteString("["), color.BlueString(format), color.WhiteString("]"))
	fmt.Printf("%v%v\n", banner, color.HiGreenString(data))
}

// BlockF print info message
func BlockF(name string, data string) {
	banner := fmt.Sprintf("%v%v%v ", color.WhiteString("["), color.GreenString(name), color.WhiteString("]"))
	fmt.Printf(fmt.Sprintf("%v%v\n", banner, data))
}

// BadBlockF print info message
func BadBlockF(name string, data string) {
	banner := fmt.Sprintf("%v%v%v ", color.WhiteString("["), color.RedString(name), color.WhiteString("]"))
	fmt.Printf(fmt.Sprintf("%v%v\n", banner, data))
}

// InforF print info message
func InforF(format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...))
}

// Infor print info message
func Infor(args ...interface{}) {
	logger.Info(args...)
}

// ErrorF print good message
func ErrorF(format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...))
}

// Error print good message
func Error(args ...interface{}) {
	logger.Error(args...)
}

// WarnF print good message
func WarnF(format string, args ...interface{}) {
	logger.Warning(fmt.Sprintf(format, args...))
}

// Warn print good message
func Warn(args ...interface{}) {
	logger.Warning(args...)
}

// TraceF print good message
func TraceF(format string, args ...interface{}) {
	logger.Trace(fmt.Sprintf(format, args...))
}

// Trace print good message
func Trace(args ...interface{}) {
	logger.Trace(args...)
}

// DebugF print debug message
func DebugF(format string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(format, args...))
}

// Debug print debug message
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Emojif print good message
func Emojif(e string, format string, args ...interface{}) string {
	emj := strings.TrimSpace(emoji.Sprint(e))
	return fmt.Sprintf("%1s %s", emj, fmt.Sprintf(format, args...))
}
