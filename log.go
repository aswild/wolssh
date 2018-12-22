package main

import (
    "fmt"
    "os"

    golog "log"
)

type LogLevel int

const (
    LOG_LEVEL_FATAL LogLevel = iota
    LOG_LEVEL_ERROR
    LOG_LEVEL_WARNING
    LOG_LEVEL_INFO
    LOG_LEVEL_DEBUG
)

// log level strings, must match order LogLevel values above
var logLevelStrings = [...]string{"FATAL", "E", "W", "I", "D"}

type WolSshLogger struct {
    logger  *golog.Logger
    level   LogLevel
}

func NewWolSshLogger() (*WolSshLogger) {
    l := new(WolSshLogger)
    l.logger = golog.New(os.Stderr, "", golog.Ldate | golog.Ltime)
    l.level = LOG_LEVEL_DEBUG
    return l
}

func (l *WolSshLogger) vlog(level LogLevel, format string, v ...interface{}) {
    if level <= l.level {
        //flags := golog.Ldate | golog.Ltime
        //if level == LOG_LEVEL_FATAL || level == LOG_LEVEL_DEBUG {
        //    // include file/line for fatal and debug logs
        //    flags = flags | golog.Lshortfile
        //}
        //l.logger.SetFlags(flags)
        msg := fmt.Sprintf("[%s] ", logLevelStrings[level]) + fmt.Sprintf(format, v...)
        l.logger.Output(3, msg)
    }
}

func (l *WolSshLogger) Fatal(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_FATAL, format, v...)
    os.Exit(1)
}

func (l *WolSshLogger) Error(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_ERROR, format, v...)
}

func (l *WolSshLogger) Warning(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_WARNING, format, v...)
}

func (l *WolSshLogger) Info(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_INFO, format, v...)
}

func (l *WolSshLogger) Debug(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_DEBUG, format, v...)
}
