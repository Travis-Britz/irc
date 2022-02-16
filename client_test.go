package irc_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/Travis-Britz/irc"
	"github.com/Travis-Britz/irc/irctest"
)

func TestClient_ConnectAndRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	server := newServer()
	defer server.Close()

	client := &irc.Client{Nickname: "HelloBot"}
	client.DialFn = func() (io.ReadWriteCloser, error) {
		return server, nil
		// return ircdebug.WriteTo(os.Stdout, server, "-> ", ""), nil
	}
	h := &irc.Router{}
	h.OnConnect(func(w irc.MessageWriter, m *irc.Message) {
		w.WriteMessage(irc.Join("#asd"))
	})
	h.OnJoin(func(w irc.MessageWriter, m *irc.Message) {
		w.WriteMessage(irc.Quit("bye"))
	}).MatchClient(client).MatchChan("#asd")

	err := client.ConnectAndRun(ctx, h)
	if err != nil {
		t.Errorf("expected client to exit without errors, got: %v", err)
	}

}

func newServer() *irctest.Server {
	s := irctest.NewServer()
	state := struct {
		servername   string
		clientPrefix irc.Prefix
		connected    bool
	}{clientPrefix: irc.Prefix{Host: "1.2.3.4"}, servername: "irc.example.com"}

	connectSuccess := func() {
		state.connected = true
		s.WriteString(fmt.Sprintf(":%s 001 %s :Welcome to the IRC Network %s\r\n", state.servername, state.clientPrefix.Nick, state.clientPrefix.String()))
		s.WriteString(fmt.Sprintf(":%s 002 %s :Your host is %s, running version 69\r\n", state.servername, state.clientPrefix.Nick, state.servername))
		s.WriteString(fmt.Sprintf(":%s 003 %s :-\r\n", state.servername, state.clientPrefix.Nick))
		s.WriteString(fmt.Sprintf(":%s 004 %s :-\r\n", state.servername, state.clientPrefix.Nick))
		s.WriteString(fmt.Sprintf("PING :9324421\r\n"))
		s.WriteString(fmt.Sprintf(":%s 396 %s %s :is now your displayed host\r\n", state.servername, state.clientPrefix.Nick, state.clientPrefix.Host))
	}

	s.Handler = irc.HandlerFunc(func(w irc.MessageWriter, m *irc.Message) {
		m.Source = state.clientPrefix

		switch m.Command {
		case "QUIT":
			s.WriteString(fmt.Sprintf("ERROR :Closing link: %s (QUIT: %s)\r\n", m.Source.Nick, m.Params.Get(1)))
			_ = s.Close()

		case "USER":
			if !state.connected {
				state.clientPrefix.User = "~" + m.Params.Get(1)
				if state.clientPrefix.Nick != "" {
					connectSuccess()
				}
			}

		case "NICK":
			newnick := irc.Nickname(m.Params.Get(1))
			if !state.connected {
				state.clientPrefix.Nick = newnick
				if state.clientPrefix.User != "" {
					connectSuccess()
				}
				return
			}
			s.WriteString(fmt.Sprintf(":%s NICK :%s", state.clientPrefix.String(), newnick))
			state.clientPrefix.Nick = newnick
		case "JOIN":
			s.WriteString(fmt.Sprintf(":%s JOIN :%s\r\n", state.clientPrefix.String(), m.Params.Get(1)))
		}

	})

	return s
}

type connStatus int

const (
	statusDisconnected connStatus = iota
	statusRegistering
	statusRegistered
	statusConnected
)
