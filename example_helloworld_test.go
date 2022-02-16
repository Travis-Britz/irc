package irc_test

import (
	"context"
	"log"

	"github.com/Travis-Britz/irc"
)

// Hello, #World:
// The following code connects to an IRC server,
// waits for RPL_WELCOME,
// then requests to join a channel called #world,
// waits for the server to tell us that we've joined,
// then sends the message "Hello!" to #world,
// then disconnects with the message "Goodbye.".
func Example() {
	bot := &irc.Client{
		Addr:     "irc.example.com:6697",
		Nickname: "HelloBot",
	}
	r := &irc.Router{}
	r.OnConnect(func(w irc.MessageWriter, m *irc.Message) {
		w.WriteMessage(irc.Join("#world"))
	})
	r.OnJoin(func(w irc.MessageWriter, m *irc.Message) {
		w.WriteMessage(irc.Msg("#world", "Hello!"))
		w.WriteMessage(irc.Quit("Goodbye."))
	}).MatchChan("#world").MatchClient(bot)

	// run the bot (blocking until exit)
	err := bot.ConnectAndRun(context.Background(), r)
	if err != nil {
		log.Println(err)
	}
}
