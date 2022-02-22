This is a work in progress.
The API is highly unstable.

# IRC

Package `irc` provides an IRC client implementation for Go (golang).

## Features

- Extensive [documentation](https://pkg.go.dev/github.com/Travis-Britz/irc)
- A small, but powerful API
- Handler packages for common functionality
- Useful out of the box

## Usage Example: Hello, #World

The following code connects to an IRC server,
waits for RPL_WELCOME,
then requests to join a channel called #world,
waits for the server to tell us that we've joined,
then sends the message "Hello!" to #world,
then disconnects with the message "Goodbye.".

```go
package main

import (
  "context"
  "log"

  "github.com/Travis-Britz/irc"
)

func main() {
  bot := &irc.Client{
    Addr: "irc.example.com:6697",
    Nickname: "HelloBot",
  }

  handler := &irc.Router{}
  handler.OnConnect(func(w irc.MessageWriter, m *irc.Message) {
    w.WriteMessage(irc.Join("#world"))
  })
  handler.OnJoin(func(w irc.MessageWriter, m *irc.Message) {
    w.WriteMessage(irc.Msg("#world", "Hello!"))
    w.WriteMessage(irc.Quit("Goodbye."))
  }).MatchChan("#world").MatchClient(bot)

  err := bot.ConnectAndRun(context.Background(), handler)
  if err != nil {
    log.Println(err)
  }
}
```

More detailed examples are available in the [examples](https://pkg.go.dev/github.com/Travis-Britz/irc#pkg-examples) section of the godoc.

## Docs TODO

These are quick descriptions for concepts which need to be written and expanded with more detail.

* Extensions which add support for capabilities should listen for the CAP LS event to write their CAP REQ, and should
  listen for CAP ACK/NAK/LIST before enabling themselves. LS and LIST will always have the list of capabilities in the
  final parameter.
* Extensions should attach themselves as middleware, and almost always call the next handler. Failing to do so may break
  the client.
* Route matchers, handler ordering rules

### Decisions TODO

Things to decide that will be hard to change later without breaking code.

* Should the router stop on the first matching route or should it run every matching route?
Running all would be less complicated for users to figure out and less of a headache when commands aren't working.
It would also force them to be more careful to only let one route match.
* Should the client handle all capability negotiation (and extensions register their ability to handle caps with a callback) or should the client just pass caps on and expect each extension to do its own capability message parsing? Doing everything with extension middleware is easier to code for me, but perhaps less reliable in terms of quality if every extension author is expected to understand capability negotiation as well as their specific extension's role.

  

## FAQ

1. Can I use this in Production? Yes
1. What is NAQ? Never Asked Questions

## NAQ

1. Should I use this in Production? No, lmao

### Reference link dump

- https://modern.ircdocs.horse/
- https://tools.ietf.org/html/rfc1459#section-4.2.1
- https://tools.ietf.org/html/rfc2812#section-3.1.3
- https://tools.ietf.org/id/draft-oakley-irc-ctcp-01.html
- http://www.irc.org/tech_docs/draft-brocklesby-irc-isupport-03.txt
- https://ircv3.net/irc/
- https://ircv3.net/specs/extensions/message-tags
- https://www.mirc.com/help/mirc.html
