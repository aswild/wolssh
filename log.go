/*******************************************************************************
* log.go: homemade logging, will probably be split into a module someday
*
* Copyright 2018 Allen Wild <allenwild93@gmail.com>
* SPDX-License-Identifier: MIT
*******************************************************************************/

package main

import (
    "fmt"
    "log/syslog"
    "os"
    "sync"
    "time"
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

type Logger struct {
    Level       LogLevel
    Timestamp   bool
    Stderr      bool
    logfile     *os.File
    syslog      *syslog.Writer
    mtx         sync.Mutex
}

type SyslogConfig struct {
    facility int
    tag      string
}

func (l *Logger) Close() {
    l.mtx.Lock()
    defer l.mtx.Unlock()
    if l.syslog != nil {
        l.syslog.Close()
        l.syslog = nil
    }
    if l.logfile != nil {
        l.logfile.Close()
        l.logfile = nil
    }
}

// Open a new log file. The existing one will be closed.
func (l *Logger) SetLogFile(logfile string) {
    l.mtx.Lock()
    defer l.mtx.Unlock()
    l.setLogFile(logfile)
}
func (l *Logger) setLogFile(logfile string) {
    if l.logfile != nil {
        l.logfile.Close()
        l.logfile = nil
    }

    if logfile != "" {
        f, err := os.OpenFile(logfile, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0644)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to open log file: %s\n", err)
            os.Exit(1)
        }
        l.logfile = f
    }
}

// set the syslog facility
func (l *Logger) SetSyslog(slc *SyslogConfig) {
    l.mtx.Lock()
    defer l.mtx.Unlock()
    l.setSyslog(slc)
}
func (l *Logger) setSyslog(slc *SyslogConfig) {
    if l.syslog != nil {
        l.syslog.Close()
        l.syslog = nil
    }

    if slc != nil {
        pri := syslog.Priority(slc.facility << 3) | syslog.LOG_NOTICE
        sl, err := syslog.New(pri, slc.tag)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to connect to syslog: %s\n", err)
        } else {
            l.syslog = sl
        }
    }
}

func (l *Logger) vlog(level LogLevel, format string, v ...interface{}) {
    if level <= l.Level {
        l.mtx.Lock()
        defer l.mtx.Unlock()

        // combine the log level code with the printf args so there's
        // only one fmt.Sprintf call.
        fmtargs := append([]interface{}{logLevelStrings[level]}, v...)
        msg := fmt.Sprintf("[%s] " + format + "\n", fmtargs...)
        var tsmsg string
        if l.Timestamp {
            tsmsg = time.Now().Format("2006-01-02 15:04:05") + " " + msg
        } else {
            tsmsg = msg
        }

        if l.Stderr {
            os.Stderr.WriteString(tsmsg)
        }
        if l.logfile != nil {
            l.logfile.WriteString(tsmsg)
        }
        if l.syslog != nil {
            switch level {
                case LOG_LEVEL_FATAL:
                    l.syslog.Crit(msg)
                case LOG_LEVEL_ERROR:
                    l.syslog.Err(msg)
                case LOG_LEVEL_WARNING:
                    l.syslog.Warning(msg)
                case LOG_LEVEL_INFO:
                    l.syslog.Info(msg)
                case LOG_LEVEL_DEBUG:
                    l.syslog.Debug(msg)
            }
        }
    }
}

func (l *Logger) Fatal(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_FATAL, format, v...)
    os.Exit(1)
}

func (l *Logger) Error(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_ERROR, format, v...)
}

func (l *Logger) Warning(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_WARNING, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_INFO, format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
    l.vlog(LOG_LEVEL_DEBUG, format, v...)
}
