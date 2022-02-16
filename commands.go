package irc

// Msg constructs a new Message of type PRIVMSG,
// with target being the intended target channel or nickname,
// and message being the text body.
func Msg(target, message string) *Message {
	return NewMessage(CmdPrivmsg, target, message)
}

// Notice constructs a new message of type NOTICE,
// with target being the intended target channel or nickname,
// and message being the text body.
func Notice(target, message string) *Message {
	return NewMessage(CmdNotice, target, message)
}

// Describe constructs a new Message of type CTCP ACTION,
// with target being the intended target channel or nickname,
// and message being the text body.
//
// Describe is equivalent to the "/me" or "/describe" commands that one might enter into the text input field of popular IRC clients.
//
// By convention, actions are written in third-person.
//
// Actions are often displayed with different formatting from regular messages.
// It is common for clients to display actions with italicised text and use a different color,
// and sometimes prefix the message with an asterisk followed by the user's nickname.
// The specific display formatting varies depending on which client program each user is connecting with.
//
// For example, compare an action message with a regular privmsg:
//
//  Describe("#foo", "slaps Bob around a bit with a large trout")
//  Msg("#foo", "take that!")
//
// is equivalent to typing
//
//  /me slaps Bob around a bit with a large trout
//  take that!
//
// in channel #foo on most IRC clients, and might be displayed by a receiving client as
//
//  * Alice slaps Bob around a bit with a large trout
//  <Alice> take that!
//
// but with italics and possibly colorized.
//
func Describe(target, action string) *Message {
	return CTCP(target, "ACTION", action)
}

// TagMsg constructs a TAGMSG command, defined in the IRCv3 message-tags capability.
func TagMsg(tags map[string]string) *Message {
	return &Message{
		Tags:    tags,
		Command: CmdTagMsg,
	}
}

// CTCP constructs a CTCP (Client-to-Client Protocol) encoded
// message to the target. command is the CTCP subcommand.
func CTCP(target, command, message string) *Message {
	return NewMessage(CmdPrivmsg, target, "\x01"+command+" "+message+"\x01")
}

// CTCPReply constructs a message encoded in the CTCP reply format.
// target should be the nickname that sent us a CTCP message,
// command is the subcommand that was sent to us,
// and message depends on the type of query.
func CTCPReply(target, command, message string) *Message {
	return NewMessage(CmdNotice, target, "\x01"+command+" "+message+"\x01")
}

// Nick constructs a nickname change command.
func Nick(name string) *Message {
	return NewMessage(CmdNick, name)
}

// Join constructs a channel join command.
func Join(channel string) *Message {
	return NewMessage(CmdJoin, channel)
}

// JoinWithKey constructs a channel join command for channels that require a key (channel mode +k is set).
func JoinWithKey(channel, key string) *Message {
	return NewMessage(CmdJoin, channel, key)
}

// Part constructs leave (depart) command for channel.
func Part(channel string) *Message {
	return NewMessage(CmdPart, channel)
}

// PartWithReason is the same as Part, but with a message
// that may be shown to other clients
func PartWithReason(channel, reason string) *Message {
	return NewMessage(CmdPart, channel, reason)
}

// PartAll constructs a command to leave all channels.
func PartAll() *Message {
	// "JOIN 0" is a special case defined in the protocol for leaving all channels
	// https://tools.ietf.org/html/rfc2812#section-3.2.1
	return NewMessage(CmdJoin, "0")
}

// Quit constructs a command that will cause the server to terminate the client's connection,
// and may display the quit message to clients that are configured to show quit messages.
func Quit(message string) *Message {
	return NewMessage(CmdQuit, message)
}

// Kick constructs a command to kick another user from a channel.
func Kick(channel, nick string) *Message {
	return NewMessage(CmdKick, channel, nick)
}

// KickWithReason is similar to Kick, but the kick message
// will display reason.
func KickWithReason(channel, nick, reason string) *Message {
	return NewMessage(CmdKick, channel, nick, reason)
}

// Mode constructs a command to change a mode on a channel or on our client connection.
func Mode(target, flag, flagParam string) *Message {
	return NewMessage(CmdMode, target, flag, flagParam)
}

// ModeQuery constructs a command to get the current modes of target.
func ModeQuery(target string) *Message {
	return NewMessage(CmdMode, target)
}

// Invite constructs a command to invite nick to channel.
func Invite(nick, channel string) *Message {
	return NewMessage(CmdInvite, nick, channel)
}

// Ping constructs a command to PING the connection.
// The server will typically respond with PONG <message>,
// although it is possible on some networks to ping a specific server,
// in which case the original message is not returned.
//
// Ping is not the same as a CTCP ping,
// which is sent to a client or channel via a PRIVMSG command instead.
// To build a CTCP ping, use CTCP(<target>, "PING", time.Now()).
// Replies will match a Message of type CTCPReply(<yournick>, "PING", <sent timestamp>).
func Ping(message string) *Message {
	return NewMessage(CmdPing, message)
}

// Pong builds the reply to a PING from the connection.
// The reply message must be the same as the original
// PING message.
func Pong(reply string) *Message {
	return NewMessage(CmdPong, reply)
}

// CapLS requests a list of the capabilities supported
// by the server.
//
// version is the capability negotiation protocol version,
// e.g. "302" for version 3.2.
func CapLS(version string) *Message {
	return Cap("LS", version)
}

// CapReq requests capability cap be enabled for the
// client's connection.
func CapReq(cap string) *Message {
	return Cap("REQ", cap)
}

// CapList requests a list of the capabilities which
// have been negotiated and enabled for
// the client's connection.
func CapList() *Message {
	return Cap("LIST")
}

// CapEnd ends the capability negotiation.
func CapEnd() *Message {
	return Cap("END")
}

// Cap sends a CAP command as part of capability negotiation.
// args are the subcommand and parameters of the CAP command.
func Cap(args ...string) *Message {
	return NewMessage(CmdCap, args...)
}

// User is used at the beginning of a connection to specify
// the username and realname of a new user.
//
// realname may contain spaces.
//
// https://tools.ietf.org/html/rfc2812#section-3.1.3
func User(user, realname string) *Message {
	// The second param (mode) is typically not useful.
	// The third param is unused.
	// Sending "0" and "*" is specifically recommended by at least
	// one modern IRC overview, and is what mIRC does.
	return NewMessage(CmdUser, user, "0", "*", realname)
}

// Pass specifies the connection password.
func Pass(password string) *Message {
	return NewMessage(CmdPass, password)
}
