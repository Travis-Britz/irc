package irctest

import (
	"bufio"
	"encoding"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/Travis-Britz/irc"
)

// NewServer creates a new mock irc server that implements io.ReadWriteCloser.
// Don't forget to close.
func NewServer() *Server {
	s := &Server{}
	s.sendReader, s.sendWriter = io.Pipe()
	s.recvReader, s.recvWriter = io.Pipe()

	s.recv = make(chan []byte, 1)

	// should exit when Close() is called
	go s.read()
	go s.write()
	return s
}

type Server struct {
	Handler irc.Handler

	rs   sync.Once
	recv chan []byte

	recvReader *io.PipeReader
	recvWriter *io.PipeWriter

	sendReader *io.PipeReader
	sendWriter *io.PipeWriter
}

// Read is how the client reads lines from the server
func (s *Server) Read(p []byte) (int, error) {
	return s.sendReader.Read(p)
}

// Write is how a client sends messages to the server
func (s *Server) Write(p []byte) (int, error) {
	s.recv <- p
	return len(p), nil
}

func (s *Server) Close() error {
	_ = s.recvWriter.Close()
	_ = s.sendWriter.Close()
	s.rs.Do(func() {
		close(s.recv)
	})
	return nil
}

// WriteString sends messages to the client.
func (s *Server) WriteString(str string) {
	if !strings.HasSuffix(str, "\r\n") {
		str = str + "\r\n"
	}
	if _, err := s.sendWriter.Write([]byte(str)); err != nil {
		log.Println("mock server write error:", err)
	}
}

// WriteMessage sends messages from the server to the client
func (s *Server) WriteMessage(m encoding.TextMarshaler) {
	if b, err := m.MarshalText(); err != nil {
		if _, err := s.sendWriter.Write(b); err != nil {
			log.Println("mock server write error:", err)
		}
	} else {
		log.Println("marshaler:", err)
	}
}

func (s *Server) read() {
	scanner := bufio.NewScanner(s.recvReader)

	for scanner.Scan() {
		line := scanner.Bytes()
		m := new(irc.Message)
		m.IncludePrefix()
		if err := m.UnmarshalText(line); err != nil {
			log.Println("unmarshaling error:", err)
			continue
		}
		s.Handler.SpeakIRC(s, m)
	}
}

func (s *Server) write() {
	for b := range s.recv {
		if _, err := s.recvWriter.Write(b); err != nil {
			log.Println("server mock write error:", err)
		}
	}
}
