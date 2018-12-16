package main

import (
    "fmt"
    "io"
    "io/ioutil"
    "net"
    "os"
    "path/filepath"

    "golang.org/x/crypto/ssh"
)

type Server struct {
    config      ssh.ServerConfig
    pubKeys     []ssh.PublicKey
}

func NewServer() (Server) {
    s := Server{
        config: ssh.ServerConfig{
            PublicKeyCallback: keyAuthCallback,
        },
    }
    return s
}

func (s *Server) LoadHostKeys(keyDir string) {
    var keyTypes = [...]string{"rsa", "dsa", "ecdsa"}

    foundKey := false
    for _, t := range keyTypes {
        keyName := "ssh_host_" + t + "_key"
        keyPath := filepath.Join(keyDir, keyName)
        if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
            keyData, err := ioutil.ReadFile(keyPath)
            if err != nil {
                log.Fatal("Failed to read file '%s': %s", keyPath, err)
            }

            key, err := ssh.ParsePrivateKey(keyData)
            if err != nil {
                log.Fatal("Failed to parse private key from file '%s': %s", keyPath, err)
            }

            s.config.AddHostKey(key)
            foundKey = true
            log.Debug("Added server host key type %s from %s ", t, keyPath)
        }
    }

    if !foundKey {
        log.Fatal("Couldn't find any host keys in directory '%s'", keyDir)
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
            log.Debug("Connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

            go ssh.DiscardRequests(reqs)
            go handleChannels(chans)
        }()
    }
}

func keyAuthCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
    log.Debug("Auth connection from %s with public key type %s", conn.RemoteAddr(), key.Type())
    return nil, nil
}

func handleChannels(chans <-chan ssh.NewChannel) {
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

        go handleChannelRequest(channel, requests)
    }
}

func handleChannelRequest(channel ssh.Channel, reqs <-chan *ssh.Request) {
    defer channel.Close()
    for req := range reqs {
        ok := false
        switch req.Type {
            case "exec":
                cmd := string(req.Payload[4:4+req.Payload[3]])
                log.Info("request to execute command '%s'", cmd)
                io.WriteString(channel, fmt.Sprintf("You requested to execute command '%s'\n", cmd))
                ok = true

            case "shell":
                log.Info("request shell")
                io.WriteString(channel, "You requested a shell\n")
                ok = true

            default:
                log.Info("request other channel type: %s", req.Type)
        }
        req.Reply(ok, nil)
        channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
        if ok {
            return
        }
    }
}
