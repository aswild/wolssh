/*******************************************************************************
* server.go: SSH Server internals
*
* Copyright 2018 Allen Wild <allenwild93@gmail.com>
* SPDX-License-Identifier: MIT
*******************************************************************************/

package main

import (
    "fmt"
    "io"
    "io/ioutil"
    "net"
    "path/filepath"
    "reflect"

    "golang.org/x/crypto/ssh"
)

type Server struct {
    config      ssh.ServerConfig
    userKeys    map[string]map[string]string
}

func NewServer() (*Server) {
    s := Server{
        config:     ssh.ServerConfig{},
        userKeys:   map[string]map[string]string{},
    }
    s.config.PublicKeyCallback = s.authPublicKey

    return &s
}

func (s *Server) authPublicKey(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
    user := conn.User()
    if keys, ok := s.userKeys[user]; ok {
        if comment, ok := keys[string(pubKey.Marshal())]; ok {
            return &ssh.Permissions{
                Extensions: map[string]string{
                    "pubkey-fp": ssh.FingerprintSHA256(pubKey),
                    "pubkey-comment": comment,
                },
            }, nil
        }
        return nil, fmt.Errorf("connection from %v: unknown public key for %q", conn.RemoteAddr(), user)
    }
    return nil, fmt.Errorf("connection from %v: unknown user %q", conn.RemoteAddr(), user)
}

func (s *Server) LoadHostKeys(paths []string) {
    // key types we've found (map to avoid duplicates)
    foundKeys := map[string]bool{}
    for _, p := range paths {
        globs, _ := filepath.Glob(p)
        for _, keyPath := range globs {
            keyData, err := ioutil.ReadFile(keyPath)
            if err != nil {
                log.Error("Failed to read host key: %s", err)
                continue
            }

            key, err := ssh.ParsePrivateKey(keyData)
            if err != nil {
                log.Error("Failed to parse private key '%s': %s", keyPath, err)
            }

            s.config.AddHostKey(key)
            foundKeys[key.PublicKey().Type()] = true
            log.Debug("Loaded host key %s", keyPath)
        }
    }

    if len(foundKeys) == 0 {
        log.Fatal("Couldn't find any host keys!")
    }

    // reflect is the only non-loopy way to get a list of keys from a map,
    // and even then it returns []reflect.Value rather than a string slice
    // (and there's no comprehension to compactly convert []Value to []string)
    log.Info("Loaded SSH host keys: %v", reflect.ValueOf(foundKeys).MapKeys())
}

func (s *Server) AddUser(name string, keys []string) {
    log.Debug("Adding user %q with %d keys", name, len(keys))
    keyMap := make(map[string]string)

    for _, key := range keys {
        data := []byte(key)
        for len(data) > 0 {
            pubKey, comment, _, rest, err := ssh.ParseAuthorizedKey(data)
            if err != nil {
                log.Fatal("failed to parse key for user %q: %v", name, err)
            }
            log.Debug("Loaded authorized public key for user %q: %q", name, comment)
            keyMap[string(pubKey.Marshal())] = comment
            data = rest
        }
    }

    if len(keyMap) == 0 {
        log.Warning("No authorized keys for user %q", name)
    } else {
        log.Info("Loaded %d authorized keys for user %q", len(keyMap), name)
    }

    if s.userKeys[name] != nil {
        log.Warning("Duplicate user %q, overwriting keys", name)
    }
    s.userKeys[name] = keyMap
}

func (s *Server) AddUsers(users []UserConfig) {
    for _, u := range users {
        s.AddUser(u.Name, u.Keys)
    }
}

func (s *Server) Listen(listenAddr string) {
    socket, err := net.Listen("tcp", listenAddr)
    if err != nil {
        log.Fatal("Failed to listen on socket: %s", err)
    }

    log.Info("listening on %s", listenAddr)
    for {
        conn, err := socket.Accept()
        if err != nil {
            log.Debug("Error accepting connection: %v", err)
            continue
        }

        go func() {
            sshConn, chans, reqs, err := ssh.NewServerConn(conn, &s.config)
            if err != nil {
                log.Error("SSH Handshake error: %v", err)
                return
            }
            //log.Debug("Connection from %q for user %q with key %q (%s)",
            //          sshConn.RemoteAddr(), sshConn.User(), sshConn.Permissions.Extensions["pubkey-fp"],
            //          sshConn.Permissions.Extensions["pubkey-comment"])
            log.Debug("Connection from %q for user %q with key (%s)",
                      sshConn.RemoteAddr(), sshConn.User(), sshConn.Permissions.Extensions["pubkey-comment"])

            go ssh.DiscardRequests(reqs)
            for newChannel := range chans {
                if t := newChannel.ChannelType(); t != "session" {
                    newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
                    continue
                }

                channel, requests, err := newChannel.Accept()
                if err != nil {
                    log.Error("could not accept channel: %s", err)
                    continue
                }

                go handleChannelRequests(channel, requests)
            }
        }()
    }
}

func handleChannelRequests(channel ssh.Channel, reqs <-chan *ssh.Request) {
    defer channel.Close()
    for req := range reqs {
        exitStatus := byte(1)
        ok := false
        switch req.Type {
            case "exec":
                cmd := string(req.Payload[4:4+req.Payload[3]])
                log.Info("request to execute command '%s'", cmd)
                var resp string
                resp, exitStatus = HandleWolCmd(cmd)
                io.WriteString(channel, fmt.Sprintf("%s\n", resp))
                ok = true

            case "shell":
                log.Info("request shell")
                io.WriteString(channel, "Sorry, you requested a shell, but that's not allowed.\n" )
                ok = true

            default:
                log.Info("request unknown channel type: %s", req.Type)
        }
        req.Reply(ok, nil)
        channel.SendRequest("exit-status", false, []byte{0, 0, 0, exitStatus})
        if ok {
            return
        }
    }
}
