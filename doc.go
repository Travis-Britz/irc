// comment

/*
Package irc provides an IRC client implementation.

This overview provides brief introductions for types and concepts.
The godoc for each type contains expanded documentation.

Jump to the package examples to see what writing client code looks like with this package.

The README file links to available extension packages for features such as state tracking, flood protection, and more.

API

These are the main interfaces and structs that you will interact with while using this package:

	// A Handler responds to an IRC message.
	type Handler interface {
		SpeakIRC(MessageWriter, *Message)
	}

	// A MessageWriter can write an IRC message.
	type MessageWriter interface {
		WriteMessage(encoding.TextMarshaler)
	}

	// Message represents any incoming our outgoing IRC line.
	type Message struct {

		// Tags contains any IRCv3 message tags.
		Tags    Tags

		// Source is where the message originated from.
		Source  Prefix

		// Command is the IRC verb or numeric (event type) such as PRIVMSG, NOTICE, 001, etc.
		Command Command

		// Params contains all the message parameters.
		Params  Params
	}

	// A Client manages a connection to an IRC server.
	type Client struct {
		//...
	}

	// ConnectAndRun starts the client.
	func (c *Client) ConnectAndRun(ctx context.Context, h Handler) error {
		//...
	}

Client

The Client type provides a simple abstraction around an IRC connection.
It manages reading and writing messages to the IRC connection and calls your handler for each message it parses.
It also deals with protocol concerns like ping timeouts.


Handler

This interface enables the development of handler packages.

	type Handler interface {
	  SpeakIRC(MessageWriter, *Message)
	}

Such packages may implement protocols such as IRCv3 Message IDs,
or common bot concerns like preferred/alternate nickname for disconnects, flood protection, and channel state.

Because the Handler interface for the irc package mimics the signature of the http.Handler interface,
most patterns for http middleware can also be applied to irc handlers.
Search results for phrases like "golang http middleware pattern", "adapter patterns", and "decorator patterns" should all return concepts from tutorials and blog posts that can be applied to this interface.

MessageWriter

The MessageWriter interface accepts any type that knows how to marshal itself into a line of IRC-encoded text.

Most of the time it makes sense to send a Message struct,
either by using the NewMessage function or any of the related constructors such as irc.Msg, irc.Notice, irc.Describe, etc.

However, it can also be very simple to implement yourself:

	// w is an irc.MessageWriter
	w.WriteMessage(rawLine("PRIVMSG #World :Hello!"))

	type rawLine string
	// MarshalText implements encoding.TextMarshaler
	func (l rawLine) MarshalText() ([]byte,error) {
		return []byte(l),nil
	}

The named Message constructors (irc.Msg, irc.Notice, etc.) should generally be preferred because they explicitly list the available parameters for each command.
This provides type safety, ordering safety, and most IDEs will provide intellisense suggestions and documentation for each parameter.


In other words:

	// prefer this:
	w.WriteMessage(irc.Msg("#world", "Hello!"))
	// instead of this:
	w.WriteMessage(irc.NewMessage(irc.CmdPrivmsg, "#world", "Hello!"))
	// and definitely instead of this:
	w.WriteMessage(rawLine(...))

Router

The Router type is an implementation of Handler.
It provides a convenient way to route incoming messages to specific handler functions by matching against message attributes like the command, source, target channel, and much more.
It also provides an easy way to apply middleware, either globally or to specific routes.
You are not required to use it, however. You can just as easily write your own message handler.

It performs a role comparable to http.ServeMux, though it is not really a multiplexer.

	r := &irc.Router{}
	r.OnText("!watchtime*", handleCommandWatchtime)
	//...
	func handleCommandWatchtime(...) {
		//...
	}

Middleware

Middleware are just handlers.
The term "middleware" applies to handlers which follow a pattern of accepting a handler as one of their arguments and returning a handler.

	// logHandler is a function that follows the middleware pattern and implements the Handler interface.
	func logHandler(next irc.Handler) irc.HandlerFunc {
		return func(w irc.MessageWriter, m *irc.Message) {
			log.Printf("parsed: %#v\n", m)
			next.SpeakIRC(w, m)
		}
	}

	func main() {
		//...
		handler := &irc.Router{}
		//...

		err := client.ConnectAndRun(..., logHandler(handler))
	}

Middleware can intercept outgoing messages by decorating the MessageWriter,
as well as call the next handler with a modified *Message.
These two abilities allow well-written packages to provide middleware that extend a client with nearly any IRC capability.

Because the ordering of received messages is important for calculating various client states,
it is generally not safe for middleware handlers to operate concurrently unless they can maintain message ordering.

Request Lifecycle

To bring it all together, this is the general sequence of events when running a client:

	- A Client's ConnectAndRun method is called and given a Handler.
	- Internally, the client wraps the provided handler with additional middleware handlers that implement core IRC features.
	- ConnectAndRun calls DialFn to connect to an IRC stream.
	- The client will begin reading lines from the stream and parse them into Message structs until the connection is closed.

Each Message parsed from the stream will result in a call to the client's handler,
which is given a MessageWriter and reference to the parsed Message struct.
Assuming that you use the package's Router type as your handler,
this is what that sequence looks like:

	- The internal client middleware that wrapped your handler will execute on order,
	calling next.SpeakIRC until they reach the Router (your handler).
	- The global middleware attached to the Router, if any, will execute in order.
	- The Router will test each route until it finds one where all conditions match.
	- If the matched route has any middleware, they will execute in order.
	- Finally, the handler provided for the route will execute.

Any of these actions could occur at any point in the chain:

	- A handler decides to write a message.
	- A handler halts execution by returning without calling the next handler.
	- A handler interprets a message to update its own internal state.
	- A handler calls the next handler with a modified version of the message.
	- A handler calls the next handler with a new MessageWriter that is decorated with a additional functions.

Message Formatting

This package does not implement message formatting.
That is to say, there are no irc.Msgf or related functions.
Formatting requirements vary widely by application.
Some applications will want to extend the formatting rules with their own replacement sequences to include IRC color formatting in replies.
Rather than implement nonstandard rules here (and force users to look up replacements),
the canonical way to write formatted replies in the style of fmt.Printf is to write your own reply helper functions.
For example:

	func replyTof(w irc.MessageWriter, m *irc.Message, format string, args ...interface{}) {
		target,_ := m.Chan()
		reply := irc.Msg(target, fmt.Sprintf(format, ...args))
		w.WriteMessage(reply)
	}

*/
package irc
