package main

import (
    "flag"
    "strings"
    "path/filepath"
)

var version string = "v0.0.0"
var log *WolSshLogger

func main() {

    pListenAddr := flag.String("port", "2222", "Listen address/port, format [host:]port")
    pSshDir := flag.String("sshdir", "ssh", "Directory containing SSH host keys and authorized_keys")
    pLogLevel := flag.Int("loglevel", int(LOG_LEVEL_INFO),
                                "Log level number: 0=fatal, 1=error, 2=warn, 3=info (default), 4=debug")
    pDebug := flag.Bool("D", false, "Enable debug logging (same as -loglevel=4)")

    flag.Parse()

    log = NewWolSshLogger()
    log.level = LogLevel(*pLogLevel);
    if *pDebug {
        log.level = LOG_LEVEL_DEBUG
    }

    listenAddr := *pListenAddr
    if !strings.Contains(listenAddr, ":") {
        listenAddr = ":" + listenAddr
    }

    server := NewServer()
    server.LoadHostKeys(*pSshDir)
    server.LoadAuthorizedKeys(filepath.Join(*pSshDir, "authorized_keys"))
    server.Listen(listenAddr)
}
