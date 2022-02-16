package irc_test

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/Travis-Britz/irc"
)

// This example uses the message router to perform more complicated message matching with an event callback style.
// Connects to an IRC server, joins a channel called "#world", sends the message "Hello!", then quits when CTRL+C is pressed.
func Example_router() {
	ctx, cancel := context.WithCancel(context.Background())

	bot := &irc.Client{
		Addr:     "irc.swiftirc.net:6697",
		Nickname: "HelloBot",
	}

	// Router implements irc.Handler and maps incoming messages (events) to a handler.
	r := &irc.Router{}

	r.OnConnect(func(w irc.MessageWriter, m *irc.Message) {
		w.WriteMessage(irc.Join("#world"))
	})

	// r.OnKick(func(w irc.MessageWriter, m *irc.Message) {
	// 	kicked, _ := m.Affected()
	// 	if !kicked.Is(bot.Nick().String()) {
	// 		return
	// 	}
	//
	// 	w.WriteMessage(irc.Msg(e.Nick().String(), "You kicked me!"))
	// })

	r.OnJoin(func(w irc.MessageWriter, m *irc.Message) {
		w.WriteMessage(irc.Msg("#world", "Hello!"))
	}).
		MatchChan("#world").
		MatchClient(bot)

	// When somebody types "!greet nickname" we respond with "Hello, nickname!".
	r.OnText("!greet &", func(w irc.MessageWriter, m *irc.Message) {
		text, _ := m.Text()
		fields := strings.Fields(text)
		channelName, _ := m.Chan()
		reply := "Hello, " + fields[1] + "!" // the second field is guaranteed to exist due to the wildcard format
		w.WriteMessage(irc.Msg(channelName, reply))
	})

	// Listen for interrupt signals (Ctrl+C) and initiate
	// a graceful shutdown sequence when one is received.
	shutdown := make(chan os.Signal)
	go func() {
		<-shutdown
		cancel()
	}()
	signal.Notify(shutdown, os.Interrupt)

	// run the bot (blocking until exit)
	err := bot.ConnectAndRun(ctx, r)
	if err != nil {
		log.Println(err)
	}
}
