package main

import (
    "flag"
    "strings"
    "path/filepath"
)

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

    log.Info("Sending test packet")
    if err := SendWol("1c:87:2c:55:89:12"); err != nil {
        log.Error("Failed to send WOL packet: %s", err)
    }

    server := NewServer()
    server.LoadHostKeys(*pSshDir)
    server.LoadAuthorizedKeys(filepath.Join(*pSshDir, "authorized_keys"))
    server.Listen(listenAddr)
}
