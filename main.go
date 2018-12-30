package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "path/filepath"
    "strings"
    "syscall"
)

var version string = "v0.0.0"
var log *Logger

var opts struct{
    showVersion bool
    debug       bool
    logLevel    int
    logFile     string
    logStderr   bool
    syslog      bool
    syslogFac   int
    syslogTag   string
    listenAddr  string
    sshDir      string
}

func main() {
    // main options
    flag.BoolVar(&opts.showVersion, "V", false, "Show version and exit")
    flag.StringVar(&opts.listenAddr, "port", "2222", "Listen address/port, format [host:]port")
    flag.StringVar(&opts.sshDir, "sshdir", "ssh", "Directory containing SSH host keys and authorized_keys")

    // log options
    flag.BoolVar(&opts.debug, "D", false, "Enable debug logging (same as -loglevel=4)")
    flag.IntVar(&opts.logLevel, "loglevel", int(LOG_LEVEL_INFO),
                "Log level number: 0=fatal, 1=error, 2=warn, 3=info (default), 4=debug")
    flag.StringVar(&opts.logFile, "logfile", "", "Log file name (default none)")
    flag.BoolVar(&opts.logStderr, "E", false, "Log to stderr (default on unless -syslog or -logfile are given)")
    flag.BoolVar(&opts.syslog, "syslog", false, "Log to syslog (default off)")
    flag.IntVar(&opts.syslogFac, "facility", 18, "syslog facility code (default 18/local2)")
    flag.StringVar(&opts.syslogTag, "tag", "wolssh", "syslog message tag")

    flag.Parse()

    if opts.showVersion {
        fmt.Println("wolssh version", version)
        os.Exit(0)
    }

    // logging setup
    level := LogLevel(opts.logLevel)
    if opts.debug {
        level = LOG_LEVEL_DEBUG
    }

    var syslogConfig *SyslogConfig = nil
    if opts.syslog {
        syslogConfig = &SyslogConfig{facility:opts.syslogFac, tag:opts.syslogTag}
    } else if opts.logFile == "" {
        // force enable stderr logging if no file or syslog given
        opts.logStderr = true
    }

    log = NewLogger(level, opts.logStderr, opts.logFile, syslogConfig)

    // signal handling
    sighupChan := make(chan os.Signal, 1)
    signal.Notify(sighupChan, syscall.SIGHUP)
    go func() {
        <-sighupChan
        log.SetLogFile(opts.logFile)
        log.Info("Caught SIGHUP, log file reopened")
    }()

    if !strings.Contains(opts.listenAddr, ":") {
        opts.listenAddr = ":" + opts.listenAddr
    }

    log.Info("Starting wolssh version %s", version)
    server := NewServer()
    server.LoadHostKeys(opts.sshDir)
    server.LoadAuthorizedKeys(filepath.Join(opts.sshDir, "authorized_keys"))
    server.Listen(opts.listenAddr)
}
