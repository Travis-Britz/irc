package irc_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Travis-Britz/irc"
	"github.com/Travis-Britz/irc/ircdebug"
)

func ExampleClient_dialFn() {
	client := &irc.Client{Nickname: "WiZ"}
	client.DialFn = func() (io.ReadWriteCloser, error) {
		return net.Dial("tcp", "irc.example.com:6667")
	}
}

func ExampleClient_dialFnDecorated() {
	client := &irc.Client{Nickname: "WiZ"}
	client.DialFn = func() (io.ReadWriteCloser, error) {
		conn, err := net.Dial("tcp", "irc.example.com:6667")
		return ircdebug.WriteTo(os.Stdout, conn, "-> ", "<- "), err
	}
}

// This example demonstrates why using the Get method of a Params type is preferable to accessing its slice index directly.
// Note the parsing behavior around missing and empty params.
// The parser only interprets syntax without understanding the semantics of a PART command.
// In other words, it does not know how many parameters a PART command has.
// Similarly, functions which interpret a PART command don't care about the protocol syntax difference between omitting a parameter or leaving it empty:
// in both cases they would only care about checking if the second param is equal to empty string.
func ExampleParams_get() {

	lines := []struct {
		raw         string
		description string
	}{{
		raw:         ":WiZ PART #foo",
		description: "PART with omitted reason",
	}, {
		raw:         ":WiZ PART #foo :",
		description: "PART with empty reason",
	}, {
		raw:         ":WiZ PART #foo :leaving now",
		description: `PART with reason "leaving now"`,
	},
	}

	m := &irc.Message{}
	for _, line := range lines {
		err := m.UnmarshalText([]byte(line.raw))
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("%s:\n", line.description)
		fmt.Printf("parsed: %#v\n", m.Params)
		fmt.Printf("get 1,2: %q, %q\n", m.Params.Get(1), m.Params.Get(2))
	}
	// Output:
	// PART with omitted reason:
	// parsed: irc.Params{"#foo"}
	// get 1,2: "#foo", ""
	// PART with empty reason:
	// parsed: irc.Params{"#foo", ""}
	// get 1,2: "#foo", ""
	// PART with reason "leaving now":
	// parsed: irc.Params{"#foo", "leaving now"}
	// get 1,2: "#foo", "leaving now"

}

// The Message returned by NewMessage does not have any tags set.
// This also includes the Message returned by the Msg, Notice, and other related functions.
//
// To attach tags for an outgoing message, simply access the Tags field and call the Set method before passing the message to a MessageWriter.
func ExampleNewMessage_attachingTags() {

	// h := irc.HandlerFunc(func(w irc.MessageWriter, m *irc.Message) {
	response := irc.Msg("#somechannel", "hello!")
	response.Tags.Set("msgid", "63E1033A051D4B41B1AB1FA3CF4B243E")
	//	w.WriteMessage(response)
	// })

}

// This example deals with client disconnects.
// It runs the connect loop for a client in a goroutine,
// doubling the time between reconnect attempts each time the client exits with an error.
func ExampleClient_ConnectAndRun_reconnect() {

	client := &irc.Client{Nickname: "HelloBot"}
	connected := make(chan bool, 1)
	handler := &irc.Router{}
	handler.OnConnect(func(w irc.MessageWriter, m *irc.Message) {
		select {
		case connected <- true:
		default:
		}
		w.WriteMessage(irc.Join("#World"))
	})

	ctx := context.Background()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		var delay time.Duration
		for {
			select {
			case <-ctx.Done():
				return
			case <-connected:
				delay = 0
			case <-time.After(delay):
				log.Println("connecting...")
				err := client.ConnectAndRun(ctx, handler)
				log.Println("connection ended:", err)
				select {
				case <-ctx.Done():
					// after a connection ends, make sure we're not supposed to exit before looping around again.
					return
				default:
					if err != nil {
						delay = delay*2 + time.Second
					}
					log.Println("reconnect delay:", delay)
				}

			}
		}
	}()

	wg.Wait()
}
