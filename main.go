package main

import (
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net"

    "golang.org/x/crypto/ssh"
)

func loadHostKey(keyPath string) (ssh.Signer) {
    keyData, err := ioutil.ReadFile(keyPath)
    if err != nil {
        panic(err)
    }

    key, err := ssh.ParsePrivateKey(keyData)
    if err != nil {
        panic(err)
    }

    return key
}

func keyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
    log.Println(conn.RemoteAddr(), "authenticate with public key type:", key.Type())
    return nil, nil
}

func main() {
    config := ssh.ServerConfig{
        PublicKeyCallback: keyAuth,
    }

    hostKeyDir := "./ssh"
    config.AddHostKey(loadHostKey(hostKeyDir + "/ssh_host_rsa_key"))

    port := "2222"
    socket, err := net.Listen("tcp", ":"+port)
    if err != nil {
        panic(err)
    }

    log.Printf("listening on port %s", port)
    for {
        conn, err := socket.Accept()
        if err != nil {
            log.Printf("Error accepting connection: %v", err)
            continue
        }

        go func() {
            sshConn, chans, reqs, err := ssh.NewServerConn(conn, &config)
            if err != nil {
                log.Printf("SSH Handshake error: %v", err)
                return
            }
            log.Printf("Connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

            go ssh.DiscardRequests(reqs)
            go handleChannels(chans)
        }()
    }
}

func handleChannels(chans <-chan ssh.NewChannel) {
    for newChannel := range chans {
        if t := newChannel.ChannelType(); t != "session" {
            newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
            continue
        }

        channel, requests, err := newChannel.Accept()
        if err != nil {
            log.Printf("could not accept channel: %s", err)
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
                log.Printf("request to execute command '%s'", cmd)
                io.WriteString(channel, fmt.Sprintf("You requested to execute command '%s'\n", cmd))
                ok = true

            case "shell":
                log.Println("request shell")
                io.WriteString(channel, "You requested a shell\n")
                ok = true

            default:
                log.Printf("request other channel type: %s", req.Type)
        }
        req.Reply(ok, nil)
        channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
        if ok {
            return
        }
    }
}
