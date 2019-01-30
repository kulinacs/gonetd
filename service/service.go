package service

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
)

type service struct {
	Name              string
	Port              int
	Command           string
	activeConnections int64
}

// Config is loaded from the config file to be handled
type Config struct {
	Service []*service
}

// ActiveConnections returns the current number of active connections
func (s *service) ActiveConnections() int64 {
	return atomic.LoadInt64(&s.activeConnections)
}

// Handle creates a socket for incoming connections
func (s *service) Handle(wg *sync.WaitGroup) {
	defer wg.Done()
	soc, err := net.Listen("tcp4", ":"+strconv.Itoa(s.Port))
	if err != nil {
		log.WithFields(log.Fields{"name": s.Name, "port": s.Port, "err": err}).Error("failed to start handler")
		return
	}
	log.WithFields(log.Fields{"name": s.Name, "port": s.Port}).Info("starting handler")
	for {
		conn, _ := soc.Accept()
		log.WithFields(log.Fields{"name": s.Name, "addr": conn.RemoteAddr()}).Info("new connection")
		atomic.AddInt64(&s.activeConnections, 1)
		go s.handleConnect(conn, conn)
	}
}

func (s *service) handleConnect(reader io.Reader, writer io.Writer) {
	defer atomic.AddInt64(&s.activeConnections, -1)
	cmd := exec.Command(s.Command)
	cmd.Stdin = reader
	cmd.Stdout = writer
	cmd.Stderr = writer
	cmd.Run()
}
