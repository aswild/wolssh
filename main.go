package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

var version string = "v0.0.0"
var log *Logger

var opts struct{
    showVersion bool
    debug       bool
    logLevel    int
    listenAddr  string
    sshDir      string
}

func main() {
    flag.BoolVar(&opts.showVersion, "V", false, "Show version and exit")
    flag.BoolVar(&opts.debug, "D", false, "Enable debug logging (same as -loglevel=4)")
    flag.StringVar(&opts.listenAddr, "port", "2222", "Listen address/port, format [host:]port")
    flag.StringVar(&opts.sshDir, "sshdir", "ssh", "Directory containing SSH host keys and authorized_keys")
    flag.IntVar(&opts.logLevel, "loglevel", int(LOG_LEVEL_INFO),
                "Log level number: 0=fatal, 1=error, 2=warn, 3=info (default), 4=debug")

    flag.Parse()

    if opts.showVersion {
        fmt.Println("wolssh version", version)
        os.Exit(0)
    }

    level := LogLevel(opts.logLevel)
    if opts.debug {
        level = LOG_LEVEL_DEBUG
    }
    //log = NewLogger(level, true, "test.log", &SyslogConfig{facility:18, tag:"wolssh"})
    log = NewLogger(level, true, "", nil)

    if !strings.Contains(opts.listenAddr, ":") {
        opts.listenAddr = ":" + opts.listenAddr
    }

    log.Info("Starting wolssh")
    server := NewServer()
    server.LoadHostKeys(opts.sshDir)
    server.LoadAuthorizedKeys(filepath.Join(opts.sshDir, "authorized_keys"))
    server.Listen(opts.listenAddr)
}
