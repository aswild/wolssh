/*******************************************************************************
 * magic_packet.go
 *
 * https://github.com/sabhiram/go-wol/blob/master/magic_packet.go
 * Retrieved 2018-12-22 from git rev 5d8c3d0e40c5501b39f110ee4dec02a8db671913
 *
 * I copied this source instead of importing it to avoid unnecessary extra
 * dependencies used by other parts of go-wol.
 *
 * Modifications by Allen Wild:
 *  - convert tabs to spaces
 *  - change package name to "main"
 *  - rename "New" function to "NewMagicPacket"
 *
 * sabhiram/go-wol is released under the MIT license:
 *
 * Copyright (c) 2015 Shaba Abhiram
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *******************************************************************************/

package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "net"
    "regexp"
)

var (
    delims = ":-"
    reMAC  = regexp.MustCompile(`^([0-9a-fA-F]{2}[` + delims + `]){5}([0-9a-fA-F]{2})$`)
)

// MACAddress represents a 6 byte network mac address.
type MACAddress [6]byte

// A MagicPacket is constituted of 6 bytes of 0xFF followed by 16-groups of the
// destination MAC address.
type MagicPacket struct {
    header  [6]byte
    payload [16]MACAddress
}

// New returns a magic packet based on a mac address string.
func NewMagicPacket(mac string) (*MagicPacket, error) {
    var packet MagicPacket
    var macAddr MACAddress

    hwAddr, err := net.ParseMAC(mac)
    if err != nil {
        return nil, err
    }

    // We only support 6 byte MAC addresses since it is much harder to use the
    // binary.Write(...) interface when the size of the MagicPacket is dynamic.
    if !reMAC.MatchString(mac) {
        return nil, fmt.Errorf("%s is not a IEEE 802 MAC-48 address", mac)
    }

    // Copy bytes from the returned HardwareAddr -> a fixed size MACAddress.
    for idx := range macAddr {
        macAddr[idx] = hwAddr[idx]
    }

    // Setup the header which is 6 repetitions of 0xFF.
    for idx := range packet.header {
        packet.header[idx] = 0xFF
    }

    // Setup the payload which is 16 repetitions of the MAC addr.
    for idx := range packet.payload {
        packet.payload[idx] = macAddr
    }

    return &packet, nil
}

// Marshal serializes the magic packet structure into a 102 byte slice.
func (mp *MagicPacket) Marshal() ([]byte, error) {
    var buf bytes.Buffer
    if err := binary.Write(&buf, binary.BigEndian, mp); err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}

