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
    pubKeysMap  map[string]string
}

func NewServer() (Server) {
    s := Server{
        config:     ssh.ServerConfig{},
        pubKeysMap: map[string]string{},
    }

    s.config.PublicKeyCallback = func(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
        if comment, ok := s.pubKeysMap[string(pubKey.Marshal())]; ok {
            return &ssh.Permissions{
                Extensions: map[string]string{
                    "pubkey-fp": ssh.FingerprintSHA256(pubKey),
                    "pubkey-comment": comment,
                },
            }, nil
        }
        return nil, fmt.Errorf("unknown public key for %q", conn.User())
    }

    return s
}

func (s *Server) LoadHostKeys(keyDir string) {
    keyTypes := [...]string{"rsa", "dsa", "ecdsa", "ed25519"}

    foundKey := false
    for _, t := range keyTypes {
        keyName := "ssh_host_" + t + "_key"
        keyPath := filepath.Join(keyDir, keyName)
        keyData, err := ioutil.ReadFile(keyPath)
        if err != nil {
            if !os.IsNotExist(err) {
                log.Error("Failed to read host key: %s", err)
            }
            continue
        }

        key, err := ssh.ParsePrivateKey(keyData)
        if err != nil {
            log.Error("Failed to parse private key '%s': %s", keyPath, err)
        }

        s.config.AddHostKey(key)
        foundKey = true
        log.Info("Loaded host key type %s", t)
    }

    if !foundKey {
        log.Fatal("Couldn't find any host keys in directory '%s'", keyDir)
    }
}

func (s *Server) LoadAuthorizedKeys(keyFile string) {
    data, err := ioutil.ReadFile(keyFile)
    if err != nil {
        log.Fatal("Failed to read authorized_keys file '%s'", keyFile)
    }

    count := 0
    for len(data) > 0 {
        pubKey, comment, _, rest, err := ssh.ParseAuthorizedKey(data)
        if err != nil {
            log.Fatal("Failed to parse authorized_keys file '%s': %s", keyFile, err)
        }
        log.Debug("Loaded authorized public key %s", comment)
        s.pubKeysMap[string(pubKey.Marshal())] = comment
        data = rest
        count++
    }
    if count == 0 {
        log.Warning("Didn't load any authorized public keys!")
    } else {
        log.Info("Loaded %d authorized keys", count)
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
                log.Info("request other channel type: %s", req.Type)
        }
        req.Reply(ok, nil)
        channel.SendRequest("exit-status", false, []byte{0, 0, 0, exitStatus})
        if ok {
            return
        }
    }
}
