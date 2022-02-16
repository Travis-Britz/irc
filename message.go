package irc

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"strings"
)

// warnTruncate is an error indicating that an encoded IRC message is too long. The message
// was still sent to the server, but the server is likely to truncate the end of the
// message before sending it to other clients.
//
// Most IRC servers limit messages to 512 bytes in length, including the trailing CR-LF characters.
// Implementations which include message tags allow an additional 512 bytes for the tags
// section of a message, including the leading '@' and trailing space character(s).
// https://modern.ircdocs.horse/#messages
//
// The message-tags capability specifies 8191 bytes for the message tags portion of a message,
// including the leading '@' and trailing SPACE.
// Clients MUST NOT send messages with tag data exceeding 4094 bytes, this includes tags with or without the client-only prefix.
// https://ircv3.net/specs/extensions/message-tags.html
//
// An important distinction is that this 512-byte limit is enforced by the server when sending
// messages to *other* connected clients, which includes the full address prefix of *our*
// client. That means the longer our nickname, user, and visible hostname are, the fewer
// bytes remain for the message parameters (contents).
//
// If you know that the server which you are connected to will accept lines which exceed the
// limit defined in the IRC protocol (512 bytes including \r\n), then it is safe
// to discard this error.
// E.g.:
//     if errors.Is(err, warnTruncate) { err = nil }
var warnTruncate = errors.New("message length exceeds IRC limit and may be truncated")

// WarnTooManyParams is an error which is returned when encoding a message with more than
// 15 parameters. RFC 2812 in particular specified 15 as the limit, and defined the
// leading ':' of the trailing parameter as optional when the trailing parameter
// was 15th in the parameter list, which could cause incorrect message parsing
// on clients which followed the RFC exactly.
//
// Generally it is recommended to accept any number of incoming parameters, but never
// send more than 15 in a message.
//
// If the server and all connected clients which will receive the message are known to
// accept more than 15 parameters in any message, then it is safe to discard
// this error.
// var WarnTooManyParams = errors.New("the message has too many parameters")

// parameterLimit is the maximum number of parameters a message may contain as defined by the protocol.
// Generally, clients should never send more than this limit but should accept any number.
const parameterLimit = 15

// NewMessage constructs a new Message to be sent on the connection
// with cmd as the verb and args as the message parameters.
//
// Only the last argument may contain SPACE (ascii 32, %x20).
// This is a limitation defined in the IRC protocol.
// Including SPACE in any other argument will
// result in undefined behavior.
//
// It is common to use '*' in place of an unused parameter. This
// has the benefit of matching all cases in situations where
// a wildcard match is allowed.
func NewMessage(cmd Command, args ...string) *Message {
	p := make(Params, len(args), parameterLimit)
	for i, a := range args {
		p[i] = a
	}
	cmd.normalize()
	return &Message{
		Command: cmd,
		Params:  p,
	}
}

// Message represents any incoming or outgoing IRC line.
//
// Background
//
// IRC is a line-delimited, text-based protocol consisting of incoming and outgoing messages.
// The terms "message", "line", or "event" might be used within this package to refer to a Message
// (although "event" usually only refers to an incoming message).
//
// A message consists of four parts: tags, prefix, verb, and params.
type Message struct {

	// Tags contains IRCv3 message tags.
	// Tags are included by the server if the message-tags capability has been negotiated.
	Tags Tags

	// Source is where the message originated from.
	// It's set by the prefix portion of an IRC message.
	//
	// Source should be left empty for messages that will be written to an IRC connection.
	// See the docs for [Message.MarshalText] for more details.
	Source Prefix

	// Command is the IRC verb or numeric such as PRIVMSG, NOTICE, 001, etc.
	// It may also sometimes be referred to as the event type.
	Command Command

	// Params contains all the message parameters.
	// If a message included a trailing component,
	// it will be included without special treatment.
	// For outgoing messages,
	// only the last parameter may contain a SPACE (ascii 32).
	// Including a space in any other parameter will result in undefined behavior.
	Params Params

	// includePrefix controls whether MarshalText will write the prefix.
	includePrefix bool
}

// MarshalText implements encoding.TextMarshaler, mainly for use with irc.MessageWriter.
func (m *Message) MarshalText() ([]byte, error) {
	/*Considerations:
	- Nickname length
	- User length
	- Host length
	- Delimiters (tag start, prefix start, param delimiter space, message delimiter \r\n)
	- (Q: does server-to-server message relaying also truncate messages?)
	- Multi-target messages:
	- - Our entire message with multiple targets needs to fit into 512 bytes (easy)
	- - If the message was directed to multiple targets, each will receive the message individually.
	- - Because the message they receive includes our prefix as well as their nickname/channel in the target,
	- - the message length is also determined by the longest target name.
	- - The simple solution is to just concatenate our estimated prefix with the total outgoing message to determine if it will be truncated,
	- - but that reduces the size of messages that we can send.
	- - (How many people are even going to bother with multi-target messages?)
	*/

	buf := bytes.NewBuffer(make([]byte, 0, 1024)) // 512 for tags, 512 for
	maxLen := 300                                 // todo: use client nick/address to calculate max message size
	var tbc int                                   // tags byte count
	var err error

	if m.Tags != nil {
		buf.WriteRune(startTags)
		for k, v := range m.Tags {
			buf.WriteString(k)
			buf.WriteRune(delimTagValue)
			buf.WriteString(escaper.Replace(v))
			buf.WriteRune(delimTag)
		}
		buf.WriteRune(delimParam)

		tbc = buf.Len()
		if tbc > 8000 { // todo: this number isn't correct
			err = fmt.Errorf("%w: message tags were %d bytes", warnTruncate, tbc)
		}
	}

	if m.includePrefix && m.Source != (Prefix{}) {
		buf.WriteRune(startPrefix)
		buf.WriteString(m.Source.String())
		buf.WriteRune(delimParam)
	}

	buf.WriteString(m.Command.String())

	for i := 0; i < len(m.Params); i++ {
		buf.WriteRune(delimParam)

		// for simplicity, always write the last param in the trailing component.
		// proper parsers should handle this normally.
		if i == len(m.Params)-1 {
			buf.WriteRune(startTrailing)
		}
		buf.WriteString(m.Params[i])
	}
	buf.WriteString("\r\n")

	if l := buf.Len() - tbc; l > maxLen {
		if err != nil {
			err = fmt.Errorf("%w, and message length is %d bytes", err, l)
		}
		err = fmt.Errorf("%w: message length is %d bytes", warnTruncate, l)
	}

	return buf.Bytes(), err
}

// UnmarshalText implements encoding.TextUnmarshaler,
// accepting a line read from an IRC stream.
// text should not include the trailing CR-LF pair.
//
// This will unmarshal an arbitrarily long sequence of bytes.
// Length limitations should be implemented at the scanner.
func (m *Message) UnmarshalText(text []byte) error {

	// go start the lexer
	l := lex(string(text))

	// re-using a message to unmarshal a new line should clear old fields
	m.Source = Prefix{}
	m.Command = ""
	m.Params = nil
	m.Tags = nil

	for {
		i := l.nextItem()
		switch i.typ {
		case itemEOF:
			return nil
		case itemError:
			return errors.New(i.val)
		case itemTagKey:
			v := l.nextItem() // type itemTagValue is *always* emitted after itemTagKey
			if i.val == "" {  // if the key was empty, skip
				continue
			}
			m.Tags.Set(i.val, unescaper.Replace(v.val))
		case itemNickname:
			m.Source.Nick = Nickname(i.val)
		case itemUser:
			m.Source.User = i.val
		case itemHost:
			m.Source.Host = i.val
		case itemCommand:
			m.Command = Command(i.val)
		case itemParam:
			m.Params = append(m.Params, i.val)
		}
	}
}

// IncludePrefix controls whether the Source field will be marshaled by MarshalText.
// todo: wording
// The Source field will be included in the encoded text for the sake of compatibility with encoding.TextUnmarshaler.
// However, the Source field should be left empty for messages which are written to an IRC connection.
// This is because [RFC 1459] states that for messages originating from a client,
// it is invalid to include any prefix other than the client's nickname.
// The RFC also instructs servers to silently discard messages which do not follow this rule.
//
// [RFC 1459]: https://datatracker.ietf.org/doc/html/rfc1459#section-2.3
// todo: rename method
// The default is to enable this setting for received messages and disable it for new messages.
// Generally this should not be needed except in the case of middleware cloning a message and passing the copy to the next handler.
func (m *Message) IncludePrefix() {
	m.includePrefix = true
}

// unescaper is a string replacer that unescapes message tag values.
var unescaper = strings.NewReplacer(
	"\\:", ";",
	"\\r", "\r",
	"\\n", "\n",
	"\\s", " ",
	"\\\\", "\\",
	"\\", "",
)

// escaper is a string replacer that escapes message tag values for transmission.
var escaper = strings.NewReplacer(
	";", "\\:",
	"\r", "\\r",
	"\n", "\\n",
	" ", "\\s",
	"\\", "\\\\",
)

// Tags represents the IRCv3 message tags for an incoming or outgoing IRC line.
type Tags map[string]string

// Set will set the tag key k with value v.
func (t *Tags) Set(k string, v string) {
	if *t == nil {
		*t = make(Tags)
	}
	(*t)[k] = v
}

// Get will get the message tag value for key. All variations of missing or empty values return
// an empty string. To check whether a message included a specific tag key, use Has.
func (t Tags) Get(key string) string {
	return t[key]
}

// Has returns true when the given key was listed in the IRCv3 message tags.
func (t Tags) Has(key string) bool {
	_, ok := t[key]
	return ok
}

// Command is an IRC command such as PRIVMSG, NOTICE, 001, etc.
//
// A command may also be known as the "verb", "event type", or "numeric".
type Command string

// String implements fmt.Stringer
func (c Command) String() string {
	return string(c)
}

// normalize will modify the command to use consistent casing.
func (c *Command) normalize() {
	*c = Command(strings.ToUpper(c.String()))
}

// is does a case-insensitive compare between two commands, which is
// useful if a command was given as a string constant.
func (c Command) is(oc Command) bool {
	return strings.EqualFold(string(c), string(oc))
}

// Prefix is the optional message (line) prefix,
// which indicates the source (user or server) of the message,
// depending on the prefix format.
//
// Example line with no prefix:
// 	PING :86F3E357
//
// Example nickname-only prefix:
// 	:Travis MODE Travis :+ixz
//
// Example "fulladdress" prefix:
// 	:NickServ!services@services.host NOTICE Travis :This nickname is registered...
//
// Example server prefix:
// 	:fiery.ca.us.SwiftIRC.net MODE #foo +nt
//
type Prefix struct {
	Nick Nickname
	User string
	Host string
}

// IsServer returns true when the message originated from a server (as opposed to a user/client).
// When true, the server name will be contained in the Host field.
func (p Prefix) IsServer() bool {
	return p.Host != "" && p.Nick == ""
}

// String implements fmt.Stringer
func (p Prefix) String() string {
	switch {
	case p.Nick == "" && p.User == "" && p.Host == "":
		return ""
	case p.Nick == "" && p.User == "":
		return p.Host
	case p.User == "":
		return p.Nick.String()
	default:
		return p.Nick.String() + "!" + p.User + "@" + p.Host
	}
}

// Params contains the slice of arguments for a message.
//
// Prefer the Get method for reading params rather than accessing the slice directly.
//
// For outgoing messages,
// only the last parameter may contain SPACE (ascii 32).
// Including SPACE in any other parameter will result in undefined behavior.
//
// If a message included a trailing component as defined in [RFC 1459],
// it will be included as a normal parameter.
//
// [RFC 1459]: https://datatracker.ietf.org/doc/html/rfc1459#section-2.3.1
type Params []string

// Get returns the nth parameter (starting at 1) from the parameters list,
// or "" (empty string) if it did not exist.
//
// Because parameters have meaning based on their position in the argument list,
// and because the meaning and position depends on which command/verb was used,
// Get does not differentiate between missing and empty parameters.
// Callers do not need to worry whether a parameter exists or not;
// they may simply check whether ordinal parameter n is equivalent to empty string.
// todo: translate that paragraph to english.
func (p Params) Get(n int) string {
	if n > len(p) || n < 1 {
		return ""
	}
	return p[n-1]
}

type Nickname string

func (n Nickname) String() string {
	return string(n)
}

// Is determines whether a nickname matches a string by using Unicode case folding.
// Equal comparison does not currently factor in
func (n Nickname) Is(other string) bool {
	return strings.EqualFold(n.String(), other)
}

// MessageWriter contains methods for sending IRC messages to a server.
type MessageWriter interface {

	// WriteMessage writes the message to the client's outgoing message queue.
	// The given encoding.TextMarshaler MUST return a byte slice which conforms to the IRC protocol.
	// If the slice does not end in "\r\n", then the sequence will be appended.
	//
	// The returned slice from the MarshalText method will be written to the connection with a single call to Write.
	// If a type implements message splitting for long messages,
	// then the entire slice must consist of multiple valid "\r\n"-delimited IRC messages.
	//
	// For example:
	//  "PRIVMSG #foo :supercalifragilisticexpi-\r\nPRIVMSG #foo :alidocious\r\n"
	//
	// It is the responsibility of the MarshalText method implementer to ensure that messages are formatted correctly,
	// and in the case of custom message splitting and continuation,
	// that flood limits are not reached.
	WriteMessage(encoding.TextMarshaler)
}

// Clone creates a deep copy of m
// mainly for use by middleware that need to do concurrent processing on a message and need to ensure that future middleware don't modify their copy
// or to pass the next handler a modified message while ensuring that any existing processing won't be affected by the change
// (clean up wording)
// func (m *Message) Clone() *Message {
// 	return m
// }
