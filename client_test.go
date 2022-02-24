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

func TestClient_pongReply(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var ponged bool
	server := irctest.NewServer()
	server.Handler = irc.HandlerFunc(func(w irc.MessageWriter, m *irc.Message) {
		if m.Command == "PONG" && m.Params.Get(1) == "123456789" {
			ponged = true
			server.Close()
		}
	})
	defer server.Close()
	go server.WriteString("PING :123456789")
	client := &irc.Client{Nickname: "bot"}
	client.DialFn = func() (io.ReadWriteCloser, error) { return server, nil }
	err := client.ConnectAndRun(ctx, nil)
	if err != io.EOF {
		t.Errorf("expected client to exit with EOF; got: %v", err)
	}
	if !ponged {
		t.Errorf("PING: client never responded with PONG")
	}

}

func TestClient_ctcpRewrite(t *testing.T) {
	client, server, done := setup()
	var action, reply bool
	defer done()
	go server.WriteString(":nick PRIVMSG bot :\x01ACTION slaps bot\x01")
	go server.WriteString(":nick NOTICE bot :\x01VERSION mIRC v6.9\x01")
	handler := irc.HandlerFunc(func(w irc.MessageWriter, m *irc.Message) {
		if m.Command == irc.CTCPAction && m.Params.Get(2) == "slaps bot" {
			action = true
		}
		if m.Command == irc.CTCPVersionReply && m.Params.Get(2) == "mIRC v6.9" {
			reply = true
		}
		if action && reply {
			done()
		}
	})
	_ = client.ConnectAndRun(context.Background(), handler)
	if !action {
		t.Errorf("expected ACTION messages to be rewritten")
	}
	if !reply {
		t.Errorf("expected VERSION reply messages to be rewritten")
	}
}

func TestClient_nickTracker(t *testing.T) {
	client, server, done := setup()
	client.Nickname = "nick1"
	tested := 0
	defer done()
	// send all three lines in one string to ensure they run in order instead of sending multiple go writestring
	go server.WriteString(":irc.example.com NOTICE nick1 :test1\r\n:nick1 NICK nick2\r\n:irc.example.com NOTICE nick2 :test2\r\n")
	handler := irc.HandlerFunc(func(w irc.MessageWriter, m *irc.Message) {
		if m.Command == "NOTICE" && m.Params.Get(2) == "test1" {
			tested++
			if !client.Nick().Is("nick1") {
				t.Errorf("expected client nickname to match nick1; got %q", client.Nick())
			}
			return
		}
		if m.Command == "NOTICE" && m.Params.Get(2) == "test2" {
			defer done()
			tested++
			if !client.Nick().Is("nick2") {
				t.Errorf("expected client to report nickname as %q; got %q", "nick2", client.Nick())
			}
			return
		}
	})
	_ = client.ConnectAndRun(context.Background(), handler)
	if tested != 2 {
		t.Errorf("expected 2 tests to be run; only %d ran", tested)
	}
}

func TestNewCTCPCmd(t *testing.T) {
	fn := irc.NewCTCPCmd("ACTION")
	if irc.CTCPAction != fn {
		t.Errorf("expected NewCTCPCmd to match CTCPAction constant; got %q and %q", irc.CTCPAction, fn)
	}
}

func TestNewCTCPReply(t *testing.T) {
	fn := irc.NewCTCPReplyCmd("VERSION")
	if irc.CTCPVersionReply != fn {
		t.Errorf("expected NewCTCPCmd to match CTCPAction constant; got %q and %q", irc.CTCPVersionReply, fn)
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

func setup() (client *irc.Client, server *irctest.Server, done context.CancelFunc) {
	server = irctest.NewServer()
	client = &irc.Client{Nickname: "bot"}
	client.DialFn = func() (io.ReadWriteCloser, error) {
		return server, nil
	}
	var ctx context.Context
	ctx, done = context.WithTimeout(context.Background(), 1*time.Second)
	go func() { <-ctx.Done(); done(); server.Close() }()
	return
}
