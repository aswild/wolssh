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
var conf *Config
var log Logger = Logger{
    Level:      3,
    Timestamp:  true,
    Stderr:     true,
}

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
    flag.BoolVar(&opts.debug, "D", false, "Enable debug logging")

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
    log.Timestamp = conf.Log.Timestamp
    if opts.debug {
        conf.Log.Level = int(LOG_LEVEL_DEBUG)
    } else {
        log.Level = LogLevel(conf.Log.Level)
    }

    if conf.Log.Syslog {
        log.SetSyslog(conf.Log.Facility, conf.Log.Tag)
    } else if conf.Log.File != "" {
        log.SetLogFile(conf.Log.File)

        // SIGHUP handler to reopen the log file (e.g. after rotation)
        sighupChan := make(chan os.Signal, 1)
        signal.Notify(sighupChan, syscall.SIGHUP)
        go func() {
            <-sighupChan
            log.SetLogFile(conf.Log.File)
            log.Info("Caught SIGHUP, log file reopened")
        }()
    } else {
        // force enable stderr logging if no file or syslog given
        log.Stderr = true
    }

    // parse and verify WOL broadcast addresses
    conf.bcastAddrs = make([]BroadcastAddr, len(conf.BcastStrs))
    for i, bs := range conf.BcastStrs {
        b := MakeBroadcastAddr(bs)
        if b != nil {
            conf.bcastAddrs[i] = *b
        } else {
            log.Fatal("Invalid Broadcast address: %v", bs)
        }
    }

    if !strings.Contains(conf.Listen, ":") {
        conf.Listen = ":" + conf.Listen
    }

    log.Info("Starting wolssh version %s", version)
    server := NewServer()
    server.LoadHostKeys(conf.HostKeys)
    server.AddUsers(conf.Users)
    server.Listen(conf.Listen)
}
