package irc_test

import (
	"encoding"
	"testing"

	"github.com/Travis-Britz/irc"
)

var discard = discarder{}

type discarder struct{}

func (d discarder) WriteMessage(marshaler encoding.TextMarshaler) {}

func TestRouter_Handle(t *testing.T) {
	var callCount int
	h := func(w irc.MessageWriter, m *irc.Message) {
		callCount++
	}
	r := &irc.Router{}
	r.HandleFunc(irc.CmdPrivmsg, h)
	r.HandleFunc(irc.CmdNotice, h)

	m := irc.Msg("#foo", "!test does this work")
	r.SpeakIRC(discard, m)
	if callCount != 1 {
		t.Errorf("expected handler to be callCount once; callCount %v times", callCount)
	}
}

func TestRouter_OnText(t *testing.T) {

	tt := []struct {
		name     string
		wildcard string
		pass     []string
		fail     []string
	}{{
		"match anything",
		"*",
		[]string{"a", "*", "!foo", "!bar", "", " "},
		[]string{},
	}, {
		"match anything starting with !",
		"!*",
		[]string{"!", "!foo", "! ", "!foo bar", "!boo"},
		[]string{"", "foo!", "?foo", "f!oo"},
	}, {
		"match literal ampersand at end of word",
		"!foo&",
		[]string{"!foo&"},
		[]string{"", "!foop", "!foo &", "!foo bar"},
	}, {
		"match literal ampersand at front of word",
		"&foo&",
		[]string{"&foo&"},
		[]string{"", "!foop", "!foo &", "!foo bar", "foo foo bar"},
	}, {
		"ampersand matches word",
		"& foo &",
		[]string{"foo foo bar", "well foo kme", "!bar foo bar", "& foo &"},
		[]string{"", "!foop", "!foo &", "!foo bar", "something foo something more"},
	}, {
		"match wildcard placed anywhere",
		"!* &",
		[]string{"!foo bar", "!bar foo", "!command     space", "!foo &", "!foo bar"},
		[]string{"", "@you hey", "foo foo bar", " !f oo"},
	}, {
		"question mark matches one character",
		"?foo",
		[]string{"!foo", "?foo", ".foo", "@foo", "*foo"},
		[]string{"", "!!foo", "??foo", "..foo", "@@foo", "**foo", "!foo ", "!foo &"},
	},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			for _, given := range tc.pass {
				called := false
				handler := func(w irc.MessageWriter, m *irc.Message) {
					called = true
				}
				router := &irc.Router{}
				router.OnText(tc.wildcard, handler)
				router.SpeakIRC(discard, irc.Msg("#foo", given))
				if !called {
					t.Errorf("expected handler to be called: %q, text: %q", tc.wildcard, given)
				}
			}
			for _, given := range tc.pass {
				called := false
				handler := func(w irc.MessageWriter, m *irc.Message) {
					called = true
				}
				router := &irc.Router{}
				router.OnText(tc.wildcard, handler)
				router.SpeakIRC(discard, irc.Notice("#foo", given))
				if called {
					t.Errorf("router matched text for NOTICE when it was supposed to only match PRIVMSG")
				}
			}
			for _, given := range tc.fail {
				called := false
				handler := func(w irc.MessageWriter, m *irc.Message) {
					called = true
				}
				router := &irc.Router{}
				router.OnText(tc.wildcard, handler)
				router.SpeakIRC(discard, irc.Msg("#foo", given))
				if called {
					t.Errorf("text matched wildcard when it was not supposed to; wildcard: %q, text: %q", tc.wildcard, given)
				}
			}
		})
	}
}
