package irc

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"
)

// A Handler responds to an IRC message.
//
// An IRC message may be any type, including PRIVMSG, NOTICE, JOIN, Numerics,
// etc. It is up to the calling function to map incoming messages/commands
// to the appropriate handler.
//
// Handlers should avoid modifying the provided Message.
type Handler interface {
	SpeakIRC(MessageWriter, *Message)
}

// The HandlerFunc type is an adapter to allow the usage of ordinary functions
// as handlers, following the same pattern as http.HandlerFunc.
type HandlerFunc func(MessageWriter, *Message)

// SpeakIRC calls f(w, m).
func (f HandlerFunc) SpeakIRC(w MessageWriter, m *Message) {
	f(w, m)
}

type middleware func(Handler) Handler

func wrap(h Handler, mw ...middleware) Handler {
	if len(mw) < 1 {
		return h
	}

	wrapped := h
	// loop in reverse to preserve middleware order
	for i := len(mw) - 1; i >= 0; i-- {
		wrapped = mw[i](wrapped)
	}

	return wrapped
}

var ctcpRegex = regexp.MustCompile("^\\x01([^ \\x01]+) ?(.*?)\\x01?$")

// ctcpHandler looks for incoming PRIVMSG or NOTICE messages that match the CTCP protocol,
// and if found, modifies the Message's Command field and strips CTCP formatting from
// the message parameters before passing the message to the next Handler.
//
// ctcpHandler MUST be called before any handlers or middleware which need to
// differentiate between regular PRIVMSG/NOTICE and CTCP messages.
func ctcpHandler(next Handler) Handler {
	return HandlerFunc(func(mw MessageWriter, m *Message) {
		if !m.Command.is(CmdPrivmsg) && !m.Command.is(CmdNotice) {
			next.SpeakIRC(mw, m)
			return
		}
		body := m.Params.Get(2)
		if len(body) == 0 {
			next.SpeakIRC(mw, m)
			return
		}
		if body[0] != 0x01 { // "\x01" is the ctcp delim
			next.SpeakIRC(mw, m)
			return
		}
		parts := ctcpRegex.FindStringSubmatch(body)
		// parts should never be nil if we made it this far, but if it is we pass it on
		// because we don't know how to deal with it
		if parts == nil {
			next.SpeakIRC(mw, m)
			return
		}
		// now we know the message is either a CTCP or CTCP Reply
		subcommand := parts[1]
		body = parts[2]

		switch m.Command {
		case CmdPrivmsg:
			m.Command = CTCPAction
			m.Command = NewCTCPCmd(subcommand)
		case CmdNotice:
			m.Command = NewCTCPReplyCmd(subcommand)
		}
		m.Params[1] = body
		next.SpeakIRC(mw, m)
	})
}

// pingMiddleware intercepts server PING messages and replies with the appropriate PONG.
func pingMiddleware(next Handler) Handler {
	return HandlerFunc(func(mw MessageWriter, m *Message) {
		if !m.Command.is(CmdPing) {
			next.SpeakIRC(mw, m)
			return
		}
		mw.WriteMessage(Pong(m.Params.Get(1)))
	})
}

type pingHandler struct {
	sync.Mutex
	expecting map[string]chan bool
	timeout   func()
}

func (ph *pingHandler) ping(ctx context.Context, mw MessageWriter, m string) {
	ph.Lock()
	defer ph.Unlock()

	if ph.expecting == nil {
		ph.expecting = make(map[string]chan bool)
	}

	// if we're already expecting a reply for the given ping then we skip sending another
	// in order to simplify the logic. having duplicate in-flight pings would not
	// be of any benefit.
	if _, exists := ph.expecting[m]; exists {
		return
	}

	ret := make(chan bool, 1)
	ph.expecting[m] = ret
	go func() {
		// we know this is the only goroutine waiting for a reply to m, so when it exits
		// for any reason we must remove the reference.
		defer func() {
			ph.Lock()
			defer ph.Unlock()
			delete(ph.expecting, m)
		}()

		select {
		case <-ret:
		case <-ctx.Done():
		case <-time.After(10 * time.Second):
			ph.timeout()
		}
	}()
	mw.WriteMessage(Ping(m))
}

func (ph *pingHandler) pongHandler(next Handler) Handler {
	return HandlerFunc(func(mw MessageWriter, m *Message) {
		if !m.Command.is(CmdPong) {
			next.SpeakIRC(mw, m)
			return
		}

		ph.Lock()
		defer ph.Unlock()

		reply := m.Params.Get(2)

		// if we were not expecting the reply, pass it on
		if _, expected := ph.expecting[reply]; !expected {
			next.SpeakIRC(mw, m)
			return
		}

		// if we were expecting the reply, intercept it and don't pass it on
		select {
		case ph.expecting[reply] <- true:
		default:
		}
	})
}

// capLSHandler listens for replies to CAP LS and completes capability negotiation.
//
// "CAP * LS * :extended-join chghost cap-notify userhost-in-names multi-prefix"
// "CAP * LS :extended-join chghost cap-notify userhost-in-names multi-prefix"
// "CAP <nick> ACK :extended-join "
// "CAP <nick> LIST * :extended-join chghost cap-notify userhost-in-names multi-prefix away-notify account-notify"
// "CAP <nick> LIST :extended-join chghost cap-notify userhost-in-names multi-prefix away-notify account-notify"
// https://ircv3.net/specs/core/capability-negotiation.html
func capLSHandler(next Handler) Handler {
	return HandlerFunc(func(mw MessageWriter, m *Message) {
		// the next handler is always called first so that other middleware which request capabilities
		// will write their message before we complete negotiation.
		next.SpeakIRC(mw, m)

		if !m.Command.is(CmdCap) {
			return
		}

		// if this is ever true then something is either wrong with the server or with our message parser
		if len(m.Params) < 3 {
			return
		}

		// the 2nd param is the CAP subcommand (LS, ACK, etc.)
		switch strings.ToUpper(m.Params.Get(2)) {

		// LS lists the capabilities supported by the server
		case "LS", "NEW":
			// the list of capabilities are in the last (trailing) parameter, separated by SPACE
			// caps := strings.Fields(m.Params.Get(len(m.Params)))
			// for _, cap := range caps {
			// 	// request cap
			// }

			// An asterisk in the 3rd param (before the CAP list) indicates there will be more lines coming
			// for the CAP LS response. If this is the last line we request a list of the caps enabled and send CAP END.
			// However, if the server does not support CAP Version 302 then multiple lines will be sent without the asterisk,
			// which will cause *each* line to trigger us to send CAP LIST and CAP END. This should be fine, since additional
			// capabilities can be requested at any time (the additional requests would be sent after cap negotiation has ended).
			// Note that we send CAP END before handling the response of CAP LIST. This is intentional, since we have
			// no reason to wait for the response.
			if m.Params.Get(3) != "*" {
				mw.WriteMessage(CapList())
				mw.WriteMessage(CapEnd())
			}
		}
	})
}
