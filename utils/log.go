package utils

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/kyokomi/emoji"
    "github.com/sirupsen/logrus"
    prefixed "github.com/x-cray/logrus-prefixed-formatter"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

var logger = logrus.New()

// InitLog init log
func InitLog(options *libs.Options) {
    dir := "/tmp/osm-log/"
    if options.LogFile == "" {
        if !FolderExists(dir) {
            os.MkdirAll(dir, 0755)
        }
        tmpFile, _ := ioutil.TempFile(dir, "osmedeus-*.log")
        options.LogFile = tmpFile.Name()
    }
    dir = filepath.Dir(options.LogFile)
    if !FolderExists(dir) {
        os.MkdirAll(dir, 0755)
    }
    f, err := os.OpenFile(options.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        logger.Errorf("error opening file: %v", err)
    }

    // defer f.Close()
    mwr := io.MultiWriter(os.Stdout, f)
    logger.SetLevel(logrus.ErrorLevel)

    logger = &logrus.Logger{
        Out:   mwr,
        Level: logrus.FatalLevel,
        Formatter: &prefixed.TextFormatter{
            ForceColors:     true,
            ForceFormatting: true,
            FullTimestamp:   true,
            TimestampFormat: "2006-01-02T15:04:05",
        },
    }

    if options.Debug == true {
        logger.SetLevel(logrus.DebugLevel)
    } else if options.Verbose == true {
        logger.SetLevel(logrus.InfoLevel)
    } else if options.Quite == true {
        logger.SetOutput(ioutil.Discard)
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

// ErrorF print good message
func ErrorF(format string, args ...interface{}) {
    logger.Error(fmt.Sprintf(format, args...))
}

// WarningF print good message
func WarningF(format string, args ...interface{}) {
    good := color.YellowString("[!]")
    fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

// DebugF print debug message
func DebugF(format string, args ...interface{}) {
    logger.Debug(fmt.Sprintf(format, args...))
}

// Emojif print good message
func Emojif(e string, format string, args ...interface{}) string {
    emj := strings.TrimSpace(emoji.Sprint(e))
    return fmt.Sprintf("%1s %s", emj, fmt.Sprintf(format, args...))
}
