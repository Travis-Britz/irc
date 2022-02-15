// comment

/*
Package irc provides an IRC client implementation.

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
	// It also satisfies the encoding.TextMarshaler interface used by MessageWriter.
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

	// A Client manages a connection and parses lines into messages.
	type Client struct {
		// ...
	}

	// ConnectAndRun connects to the IRC server and runs the client until the connection is closed,
	// calling h for each message the client parses from the connection.
	func (c *Client) ConnectAndRun(h Handler) {
		// ...
	}

Encoding and Decoding

The Message type can marshal and unmarshal itself to and from a raw line of IRC-formatted text.
If you only want IRC parsing and encoding,
you can use this type for encoding or decoding IRC messages.

Request lifecycle

todo

	- A Client's ConnectAndRun method is called and given a Handler.
	- The handler is wrapped by a few middleware handlers that implement sub-protocols such as CTCP,
	and then saved to the client.
	- ConnectAndRun calls the function in the DialFn field of its Config struct to connect to an IRC stream.
	- The client will begin reading lines from the stream and parse them into Message structs until the connection is closed.
	- Each Message parsed from the stream will result in a call to the client's handler,
	which is given an object implementing MessageWriter as well as a pointer to the parsed Message struct.
-

*/
package irc
