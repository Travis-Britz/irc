package irc

import (
	"fmt"
	"strings"
	"testing"
)

func newMessage(tags map[string]string, prefix struct{ nick, user, host string }, command Command, params []string) *Message {
	p := make(Params, 0, len(params))
	for _, pa := range params {
		p = append(p, pa)
	}
	return &Message{
		Tags: tags,
		Source: Prefix{
			Nickname(prefix.nick),
			prefix.user,
			prefix.host},
		Command: command,
		Params:  p,
	}
}

func assertMessageEquals(t *testing.T, expected *Message, got *Message) {
	assertTagsEqual(t, expected.Tags, got.Tags)
	assertPrefixEqual(t, expected.Source, got.Source)
	assertCommandEquals(t, expected.Command, got.Command)
	assertParamsEqual(t, expected.Params, got.Params)
}
func assertTagsEqual(t *testing.T, expected Tags, got Tags) {
	if len(expected) != len(got) {
		t.Errorf("maps didn't match todo rename this error idfk")
	}

	for key, want := range expected {
		k, ok := got[key]
		if !ok {
			t.Errorf("actual map doesn't contain key %q: expected: %#v, got: %#v", key, expected, got)
			continue
		}

		if want != k {
			t.Errorf("actual map value \"%s\" was not equal to expected value \"%s\" in key \"%s\"", k, want, key)
			continue
		}
	}
}
func assertPrefixEqual(t *testing.T, expected Prefix, got Prefix) {
	if expected.Nick != got.Nick || expected.User != got.User || expected.Host != got.Host {
		t.Errorf("prefix didn't match; got %q wanted %q", got, expected)
	}
}
func assertCommandEquals(t *testing.T, expected Command, got Command) {

	if !got.is(expected) {
		t.Errorf("command didn't match; got %q wanted %q", got, expected)
	}
}
func assertParamsEqual(t *testing.T, expected Params, got Params) {

	if len(got) != len(expected) {
		t.Errorf("actual slice(%#v)(%d) was not the same length as expected slice(%#v)(%d)", got, len(got), expected, len(expected))
	}

	for i, v := range got {
		if v != expected[i] {
			t.Errorf("actual slice value \"%s\" was not equal to expected value \"%s\" at index \"%d\"", v, expected[i], i)
		}
	}
}
func fromBytes(b []byte) (*Message, error) {
	m := &Message{}
	err := m.UnmarshalText(b)
	return m, err
}

func TestParseMessage(t *testing.T) {
	var tags = []struct {
		raw      string
		expected map[string]string
	}{
		{"", map[string]string{}},
		{"@ ", map[string]string{}},
		{"@; ", map[string]string{}},
		{"@;; ", map[string]string{}},
		{"@k ", map[string]string{"k": ""}},
		{"@k= ", map[string]string{"k": ""}},
		{"@k=\\ ", map[string]string{"k": ""}},
		{"@k; ", map[string]string{"k": ""}},
		{"@k-l; ", map[string]string{"k-l": ""}},
		{"@k-l=; ", map[string]string{"k-l": ""}},
		{"@k;l ", map[string]string{"k": "", "l": ""}},
		{"@k;l; ", map[string]string{"k": "", "l": ""}},
		{"@k;l=; ", map[string]string{"k": "", "l": ""}},
		{"@k;l= ", map[string]string{"k": "", "l": ""}},
		{"@k;; ", map[string]string{"k": ""}},
		{"@k=\\; ", map[string]string{"k": ""}},
		{"@k=v ", map[string]string{"k": "v"}},
		{"@k=v; ", map[string]string{"k": "v"}},
		{"@k=0 ", map[string]string{"k": "0"}},
		{"@k=0; ", map[string]string{"k": "0"}},
		{"@k=\\v; ", map[string]string{"k": "v"}},
		{"@k=\\s; ", map[string]string{"k": " "}},
		{"@k=\\s ", map[string]string{"k": " "}},
		{"@k=\\: ", map[string]string{"k": ";"}},
		{"@k=\\\\ ", map[string]string{"k": "\\"}},
		{"@k=\\r ", map[string]string{"k": "\r"}},
		{"@k=\\n ", map[string]string{"k": "\n"}},
		{"@k=1;k=2; ", map[string]string{"k": "2"}},
		{"@k=\\s\\:\\r\\n\\\\\\a\\b\\ ", map[string]string{"k": " ;\r\n\\ab"}},
		{"@u==; ", map[string]string{"u": "="}},
		{"@j== ", map[string]string{"j": "="}},
		{"@draft/bot ", map[string]string{"draft/bot": ""}},
		{"@draft/bot=someFutureValueHere=2343 ", map[string]string{"draft/bot": "someFutureValueHere=2343"}},
		{"@twitch.tv/mod ", map[string]string{"twitch.tv/mod": ""}},
		{"@+twitch.tv/foo ", map[string]string{"+twitch.tv/foo": ""}},
		{"@emoji=🧔;empty;repeat=no;empty2=;zero=0;new-line=\\r\\n;repeat=yes;quote=\"; ", map[string]string{"emoji": "🧔", "empty": "", "empty2": "", "zero": "0", "new-line": "\r\n", "quote": "\"", "repeat": "yes"}},
	}

	var prefixes = []struct {
		raw      string
		expected struct {
			nick string
			user string
			host string
		}
	}{
		{"", struct{ nick, user, host string }{"", "", ""}},
		{":Bob ", struct{ nick, user, host string }{"Bob", "", ""}},
		{":Bob  ", struct{ nick, user, host string }{"Bob", "", ""}},
		{":Bob\\Loblaw ", struct{ nick, user, host string }{"Bob\\Loblaw", "", ""}},
		{":Bob\\Loblaw!@law.blog ", struct{ nick, user, host string }{"Bob\\Loblaw", "", "law.blog"}},
		{":Bob\\Loblaw!@law/blog ", struct{ nick, user, host string }{"Bob\\Loblaw", "", "law/blog"}},
		{":Bob!BLoblaw@bob.loblaw.law.blog ", struct{ nick, user, host string }{"Bob", "BLoblaw", "bob.loblaw.law.blog"}},
		{":Bob!NoHabla!@bob.loblaw.law.blog ", struct{ nick, user, host string }{"Bob", "NoHabla!", "bob.loblaw.law.blog"}},
		{":BobNoH@bl@!B.Loblaw!@bob.loblaw.law.blog ", struct{ nick, user, host string }{"BobNoH@bl@", "B.Loblaw!", "bob.loblaw.law.blog"}}, // '@' is not allowed inside nicknames on most (all?) networks, but this provides a decent parse test
		{":irc.bob.loblaw.no.habla.es ", struct{ nick, user, host string }{"", "", "irc.bob.loblaw.no.habla.es"}},
	}

	var commands = []struct {
		raw      string
		expected Command
	}{
		{"001", RplWelcome},
		{"PRIVMSG", CmdPrivmsg},
		{"Privmsg", CmdPrivmsg},
		{"privmsg", CmdPrivmsg},
		{"privmsg", Command("PRIVMSG")},
		{"PRIVMSG", Command("privmsg")},
	}

	var params = []struct {
		raw      string
		expected []string
	}{
		{"", []string{}},
		{" ", []string{""}},
		{" :", []string{""}},
		{" ::", []string{":"}},
		{" ::p1", []string{":p1"}},
		{" :p1", []string{"p1"}},
		{" p1", []string{"p1"}},
		{" p1 p2", []string{"p1", "p2"}},
		{"  p1 p2", []string{"p1", "p2"}},
		{" p1  p2", []string{"p1", "p2"}},
		{" p1  p2 :", []string{"p1", "p2", ""}},
		{" p1  p2 : ", []string{"p1", "p2", " "}},
		{" p1  p2 : :", []string{"p1", "p2", " :"}},
		{" p1  p2 : : ", []string{"p1", "p2", " : "}},
		{" p1  p2 :p3 :p3 ", []string{"p1", "p2", "p3 :p3 "}},
		{" p1  p2 :p3  :p3 ", []string{"p1", "p2", "p3  :p3 "}},
		{" p1 p2 p3 p4 p5 p6 p7 p8 p9 p10 p11 p12 p13 p14 p15 :p16", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13", "p14", "p15", "p16"}},
		{" :" + strings.Repeat("a", 513), []string{strings.Repeat("a", 513)}}, // don't blow up for lines exceeding protocol-defined length
	}

	for _, tt := range tags {
		for _, p := range prefixes {
			for _, c := range commands {
				for _, pa := range params {
					raw := fmt.Sprintf("%s%s%s%s", tt.raw, p.raw, c.raw, pa.raw)
					m, err := fromBytes([]byte(raw))
					if err != nil {
						t.Errorf("expected no error; got %v: %q", err, raw)
					}
					assertMessageEquals(t, newMessage(tt.expected, p.expected, c.expected, pa.expected), m)
				}
			}
		}
	}
}

func TestParseErrors(t *testing.T) {
	var parseErrors = []string{
		"@badge-info=;badges=;color=#FF0000;display-name=bot;emote-sets=0,19650,300374282,472873131;user-type=",
		"@badge-info=;badges=;color=#FF0000;display-name=bot;emote-sets=0,19650,300374282,472873131;user-type= ",
		"@badge-info=;badges=;color=#FF0000;display-name=bot;emote-sets=0,19650,300374282,472873131;user-type=;",
		"@badge-info=;badges=;color=#FF0000;display-name=bot;emote-sets=0,19650,300374282,472873131;user-type=; ",
		"@badge-info=;badges=;color=#FF0000;display-name=bot;emote-sets=0,19650,300374282,472873131;user-type= :tmi.twitch.tv",
		":tmi.twitch.tv",
		":Bob! TOPIC #LawBlog :Welcome to #LawBlog, where we blah blah about Bob Loblaw's Law Blog (Bob Loblaw no habla español)",
		"@",
		"@;",
		"@=",
		"@ ",
		"@; ",
		"@;= ",
		":",
		":.",
		":. ",
		":! ",
		":!@ ",
		": ",
		" ",
	}
	for _, raw := range parseErrors {
		m, err := fromBytes([]byte(raw))
		if err == nil {
			t.Errorf("expected parse error; got err == nil. raw line: %q, parsed: %#v", raw, m)
		}
	}
}
