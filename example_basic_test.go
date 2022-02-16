package irc_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Travis-Britz/irc"
)

const myName = "HelloBot"

// myHandler is an irc.HandlerFunc.
//
// On connection success (001), it joins #MyChannel.
//
// On join events, it checks if the joining nickname matched myName and the channel matched #MyChannel
// before sending an introduction.
//
// On privmsg events check if the message target matched our name (indicating a query/DM) and the first
// word begins with "Hello" before responding with "hey there!".
func myHandler(w irc.MessageWriter, m *irc.Message) {
	switch m.Command {
	case "001":
		w.WriteMessage(rawLine("JOIN #MyChannel"))
	case "JOIN":
		if !m.Source.Nick.Is(myName) {
			return
		}
		if !strings.EqualFold("#MyChannel", m.Params.Get(1)) {
			return
		}

		w.WriteMessage(rawLine("PRIVMSG #MyChannel :Hello everybody, my name is " + myName))
	case "PRIVMSG":
		if m.Params.Get(1) == myName {
			if msgBody := m.Params.Get(2); strings.HasPrefix(msgBody, "Hello") {
				w.WriteMessage(rawLine(fmt.Sprintf("PRIVMSG %s :hey there!", m.Source.Nick)))
			}
		}
	}
}

// rawLine is an IRC-formatted message.
type rawLine string

// MarshalText implements encoding.TextMarshaler, which
// is used by irc.MessageWriter.
func (l rawLine) MarshalText() ([]byte, error) {
	return []byte(l), nil
}

// The simplest possible implementation of a Message handler.
// In this case, "simple" means it is not using package features. The code should be
// considered to be a "messy" implementation, but demonstrates how easy it is to get
// down to the protocol level, if desired.
func Example_simple() {
	bot := &irc.Client{
		Addr:     "irc.example.com:6697",
		Nickname: myName,
	}

	// we need to explicitly convert myHandler to the irc.HandlerFunc type
	err := bot.ConnectAndRun(context.Background(), irc.HandlerFunc(myHandler))
	if err != nil {
		log.Fatal(err)
	}

}
