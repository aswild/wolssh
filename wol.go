// Basic Wake on LAN client library

package main

import (
    "fmt"
    "net"
)

var aliasMap = map[string]string{
    "redacted": "00:00:00:00:00:00",
}

func ResolveAlias(a string) (string, error) {
    if mac, ok := aliasMap[a]; ok {
        return mac, nil
    } else {
        return "", fmt.Errorf("Couldn't find alias '%s'", a)
    }
}

func SendWol(mac string) (error) {
    packet, err := NewMagicPacket(mac)
    if err != nil {
        return fmt.Errorf("Failed to create magic packet: %s", err)
    }

    packetBytes, err := packet.Marshal()
    if err != nil {
        return fmt.Errorf("Failed to marshal magic packet: %s", err)
    }

    baddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:40000")
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

    log.Info("Sent magic packet to %s", mac)
    return nil
}

func HandleWolCmd(cmd string) (string, byte) {
    mac, err := ResolveAlias(cmd)
    if err != nil {
        return err.Error(), 1
    }

    if err = SendWol(mac); err != nil {
        return err.Error(), 2
    }

    return fmt.Sprintf("Woke up host %s (%s)", cmd, mac), 0
}
