/*******************************************************************************
* wol.go: Basic Wake on LAN client
*
* Copyright 2018 Allen Wild <allenwild93@gmail.com>
* SPDX-License-Identifier: MIT
*******************************************************************************/

package main

import (
    "fmt"
    "net"
    "strconv"
    "strings"

    sawol "github.com/sabhiram/go-wol"
)

type BroadcastAddr struct {
    addr string
    port int
}

const defaultPort int = 40000

var aliasMap = map[string]string{
    "redacted": "00:00:00:00:00:00",
}

func MakeBroadcastAddr(s string) (*BroadcastAddr) {
    sp := strings.Split(s, ":")
    addr := sp[0]
    if !validIPv4Bcast(addr) {
        return nil
    }

    port := defaultPort
    switch len(sp) {
        case 1:
            // no-op, handled above
        case 2:
            var err error
            port, err = strconv.Atoi(sp[1])
            if err != nil {
                return nil
            }
        default:
            return nil
    }
    return &BroadcastAddr{addr, port}
}

func (b *BroadcastAddr) Marshal() string {
    return fmt.Sprintf("%s:%d", b.addr, b.port)
}

func validIPv4Bcast(addr string) bool {
    ip := net.ParseIP(addr)
    return ip != nil &&
           ip.To4() != nil &&
           !ip.IsLoopback() &&
           !ip.IsMulticast() &&
           !ip.IsUnspecified()
}

func ResolveHost(a string) (string, error) {
    if mac, ok := conf.Hosts[a]; ok {
        return mac, nil
    } else {
        return "", fmt.Errorf("Couldn't find host '%s'", a)
    }
}

func SendWol(bcast *BroadcastAddr, mac string) (error) {
    packet, err := sawol.New(mac)
    if err != nil {
        return fmt.Errorf("Failed to create magic packet: %s", err)
    }

    packetBytes, err := packet.Marshal()
    if err != nil {
        return fmt.Errorf("Failed to marshal magic packet: %s", err)
    }

    baddr, err := net.ResolveUDPAddr("udp4", bcast.Marshal())
    if err != nil {
        return fmt.Errorf("Failed to resolve UDP broadcast address: %s", err)
    }

    conn, err := net.DialUDP("udp4", nil, baddr)
    if err != nil {
        return fmt.Errorf("Failed to dial UDP connection: %s", err)
    }
    defer conn.Close()

    n, err := conn.Write(packetBytes)
    if err != nil {
        return fmt.Errorf("Failed to send magic packet")
    } else if n != 102 {
        return fmt.Errorf("expected to send 102 bytes but sent only %d", n)
    }

    log.Info("Sent magic packet for %s to %s", mac, bcast.Marshal())
    return nil
}

func HandleWolCmd(cmd string) (string, byte) {
    mac, err := ResolveHost(cmd)
    if err != nil {
        return err.Error(), 1
    }

    for _, b := range conf.bcastAddrs {
        if err = SendWol(&b, mac); err != nil {
            return err.Error(), 2
        }
    }

    return fmt.Sprintf("Woke up host %s (%s)", cmd, mac), 0
}
