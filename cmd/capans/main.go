// Package main provides an ssh server, numbskull
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var msgs = &MsgB0rker{}

func main() {
	// TODO: Configure this bitch
	srv := NewServer()
	srv.Listen(":2222")
}

type Server struct {
	cfg *ssh.ServerConfig
	// channels []chan string
}

func NewServer() *Server {
	return &Server{
		cfg: getSSHConfig(),
	}
}

func (s *Server) Listen(addr string) error {
	log.Info("Starting SSH Server...")
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	srv, chans, req, err := ssh.NewServerConn(conn, s.cfg)
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	go ssh.DiscardRequests(req)

	for newChannel := range chans {
		channel, requests, err := newChannel.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			for {
				r := <-requests
				log.Info(fmt.Sprintf("Channel request: %q", r.Type))
				r.Reply(true, nil)
			}
		}()

		go func() {
			defer channel.Close()
			StreamInstance := msgs.Stream()
			log.Info(fmt.Sprintf("client connected: %d", StreamInstance.Id()))
			go func() {
				for {
					_, err := io.Copy(channel, StreamInstance)
					if err != nil {
						fmt.Println(err)
						break
					}
				}
			}()
			io.Copy(StreamInstance, channel)
			log.Info(fmt.Sprintf("client disconnected: %d", StreamInstance.Id()))
		}()
	}

}

func getSSHConfig() *ssh.ServerConfig {
	cfg := ssh.ServerConfig{
		NoClientAuth: true,
	}

	pKeyBytes, err := ioutil.ReadFile("id_ed25519")
	if err != nil {
		panic(err)
	}

	pKey, err := ssh.ParsePrivateKey(pKeyBytes)
	if err != nil {
		panic(err)
	}

	cfg.AddHostKey(pKey)
	return &cfg
}

type MsgB0rker struct {
	messages []*bytes.Buffer
}

func (m *MsgB0rker) Stream() *Stream {
	s := &Stream{
		id: len(m.messages),
		m:  m,
	}
	m.messages = append(m.messages, &bytes.Buffer{})
	return s
}

func (m *MsgB0rker) WriteFrom(id int, p []byte) (int, error) {
	for i := range m.messages {
		if i == id {
			continue
		}
		_, err := m.messages[i].Write(p)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (m *MsgB0rker) ReadFor(id int, p []byte) (int, error) {
	return m.messages[id].Read(p)
}

type Stream struct {
	id int
	m  *MsgB0rker
}

func (s *Stream) Read(p []byte) (int, error) {
	return s.m.ReadFor(s.id, p)
}

func (s *Stream) Write(p []byte) (int, error) {
	return s.m.WriteFrom(s.id, p)
}

func (s *Stream) Id() int {
	return s.id
}
