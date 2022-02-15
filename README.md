This is an exercise for learning Go and a work in progress. The API is highly unstable as I prototype.

# IRC

Package `irc` provides IRC message parsing and a simple client.

### Features

- Extensive documentation
- A small, but powerful API
- Handler packages for common functionality
- Useful out of the box

## Usage Example: Hello, #World

The following code connects to an IRC server, joins a channel called "#World", sends the message "Hello!", then disconnects with the message "Goodbye.".


```go
package main

import (
  "log"

  "github.com/Travis-Britz/irc"
)

func main() {
  bot := irc.NewClient(&irc.Config{
	Addr:     "irc.example.com:6697",
	Nick:     "HelloBot",
  })

  handler := &irc.Router{}
  
  handler.OnConnect(func(w irc.MessageWriter, m *irc.Message) {
	w.WriteMessage(irc.Join("#world"))
  })
  
  handler.OnJoin(func(w irc.MessageWriter, m *irc.Message) {
	w.WriteMessage(irc.Msg("#world", "Hello!"))
	w.WriteMessage(irc.Quit("Goodbye."))
  }).MatchChan("#world").MatchClient(bot)
  
  err := bot.ConnectAndRun(handler)
  if err != nil {
	log.Println(err)
  }
}
```

View the package examples section of the documentation for full examples.

### Message

The [Message]() type is used to encode and decode IRC lines.

```
var err error
msg := &irc.Message{}
msg, err = msg.UnmarshalText(raw)
```

### Handler

The `irc.Handler` interface mimics the `http.Handler` pattern.
This should make writing middleware handlers, decorators, and adapters feel very familiar.

```
type Handler interface {
  SpeakIRC(MessageWriter, *Message)
}
```

This interface enables the development of handler packages,
in the same way that there is an extensive library of handlers for http.
Such packages may implement protocols such as CTCP, IRCv3 Message IDs, and more.
Common bot concerns like preferred/alternate nickname for disconnects, flood protection, and channel state can also be implemented through middleware handlers and shared easily by projects.


### Client

The `Client` type provides a simple abstraction around an IRC connection.
It manages reading and writing messages to the IRC connection and calls your handler for each message it parses.
It also deals with protocol concerns like ping timeouts.


### Router

The `Router` type is an implementation of `irc.Handler`.
It provides a convenient way to route incoming messages to specific handler functions by matching against message attributes like the command, source, message contents, and much more.
It also provides an easy way to apply middleware patterns, either globally or to specific routes.
You are not required to use it, however. You can just as easily write your own message handler.

It performs a role comparable to `http.ServeMux`, though it is not really a multiplexer.


## Docs TODO

These are quick descriptions for concepts which need to be written and expanded with more detail.

* Middleware must execute synchronously because client state is dependent upon the ordering of received messages.
* Extensions which add support for capabilities should listen for the CAP LS event to write their CAP REQ, and should
  listen for CAP ACK/NAK/LIST before enabling themselves. LS and LIST will always have the list of capabilities in the
  final parameter.
* Extensions should attach themselves as middleware, and almost always call the next handler. Failing to do so may break
  the client.
* Route matchers, handler ordering rules

Packages to write:

- Mode parsing
- `state` channel and user state tracking (Modes, Ops, ban lists, etc.)
- `ial`,`ibl` Internal Address List, Internal Ban List
- `flood` outgoing flood protection
- `isupp` 005 RPL_ISUPPORT http://www.irc.org/tech_docs/draft-brocklesby-irc-isupport-03.txt
- Travis-Britz/`tmi` Router for twitch

### Decisions TODO

Things to decide that will be hard to change later without breaking code.

* Should the router stop on the first matching route or should it run every matching route?
Running all would be less complicated for users to figure out and less of a headache when commands aren't working.
It would also force them to be more careful to only let one route match.
* Do I want to change the messagewriter interface to add a method for writing tags? 
Outgoing tags are only going to get more common,
and right now the only way to attach them through middleware is by decorating the messagewriter and doing string parsing to insert new tags.
The http responsewriter has methods for writing headers, and tags are essentially the same concept of metadata headers.
Decorators could also do reflection to assert type Message and just access the Tags field.
* Should the Params.Get method be ordinal (start at 1) or offset? Special values like 0 could return the full string, but I could also just have a method called GetFull or something.
I could also make negative numbers start counting from the end, but that's really only useful to grab the last parameter.
Messages don't usually have variable numbers of parameters anyway except for the last one, and I already take care of that case.
* Should I move the CTCP handler to a public function and take it out of the client?
Some people might want to enable it separately, but the Router also depends on it.
Pretty much all clients use CTCP, so not being built in isn't very useful.
* Should the client handle all capability negotiation (and extensions register their ability to handle caps with a callback) or should the client just pass caps on and expect each extension to do its own capability message parsing? Doing everything with extension middleware is easier to code for me, but perhaps less reliable in terms of quality if every extension author is expected to understand capability negotiation as well as their specific extension's role.
* Where do warnings (runtime errors) go? They should go somewhere like stderr by default, because silent by default is much worse.
Do I want a list of exported variables for specific errors, or just a warning type?

  

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
