/*******************************************************************************
* main.go: Wake on LAN SSH Server initialization
*
* Copyright 2018 Allen Wild <allenwild93@gmail.com>
* SPDX-License-Identifier: MIT
*******************************************************************************/

package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"
)

var version string = "v0.0.0"
var log *Logger
var conf *Config

var opts struct {
    showVersion bool
    debug       bool
    confFile    string
}

func main() {
    // main options
    flag.BoolVar(&opts.showVersion, "V", false, "Show version and exit")
    flag.StringVar(&opts.confFile, "c", "", "Configuration file")

    // log options
    flag.BoolVar(&opts.debug, "D", false, "Enable debug logging (same as -loglevel=4)")

    flag.Parse()

    if opts.showVersion {
        fmt.Println("wolssh version", version)
        os.Exit(0)
    }

    // load the config file
    conf = DefaultConfig()
    if opts.confFile != "" {
        var err error
        conf, err = LoadConfig(opts.confFile)
        if err != nil {
            fmt.Printf("Error parsing config file: %v\n", err)
            os.Exit(1)
        }
    }

    // logging setup
    if opts.debug {
        conf.Log.Level = int(LOG_LEVEL_DEBUG)
    }

    var syslogConfig *SyslogConfig = nil
    if conf.Log.Syslog {
        syslogConfig = &SyslogConfig{facility:conf.Log.Facility, tag:conf.Log.Tag}
    } else if conf.Log.File == "" {
        // force enable stderr logging if no file or syslog given
        conf.Log.Stderr = true
    }

    log = NewLogger(LogLevel(conf.Log.Level), conf.Log.Stderr, conf.Log.File, syslogConfig)

    // signal handling
    sighupChan := make(chan os.Signal, 1)
    signal.Notify(sighupChan, syscall.SIGHUP)
    go func() {
        <-sighupChan
        log.SetLogFile(conf.Log.File)
        log.Info("Caught SIGHUP, log file reopened")
    }()

    if !strings.Contains(conf.Listen, ":") {
        conf.Listen = ":" + conf.Listen
    }

    log.Info("Starting wolssh version %s", version)
    server := NewServer()
    server.LoadHostKeys(conf.SshDir)
    server.AddUsers(conf.Users)
    server.Listen(conf.Listen)
}
