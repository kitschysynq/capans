package capans

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

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

		go acceptRequests(requests)
		log.Infof("handling messages for topic %q", newChannel.ChannelType())
		go handleMsgs(channel, newChannel.ChannelType())
	}

}

func handleMsgs(channel ssh.Channel, topic string) {
	StreamInstance := msgs.Topic(topic).Stream()
	defer channel.Close()
	defer StreamInstance.Close()
	log.Infof("client connected: %d", StreamInstance.Id())
	doneChan := make(chan struct{})
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
	close(doneChan)
	log.Info(fmt.Sprintf("client disconnected: %d", StreamInstance.Id()))
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

func acceptRequests(in <-chan *ssh.Request) {
	for r := range in {
		log.Info(fmt.Sprintf("Channel request: %q", r.Type))
		r.Reply(true, nil)
	}
}
