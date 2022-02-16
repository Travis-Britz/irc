package irc

import (
	"fmt"
	"strings"
)

// PRIVMSG
// NOTICE
// CTCP ACTION
// TOPIC
// KICK
// PART
// QUIT
// ERROR
// INVITE
// MODE

// Text returns the free-form text portion of a message for the well-known (named) IRC commands.
// An error is returned if the method is called for unsupported message types.
// If err is not nil, then Text will contain the entire parameter list joined together as one string.
// However, for commands that return an error, it may be better to call Params.Get directly.
//
// Supported commands include PRIVMSG, NOTICE, PART, QUIT, ERROR, and more.
//
// In the case of PART and KICK, Text contains the <reason> message parameter.
//
// The error may be discarded without checking
// If it's known that the message will always be a supported command,
// for example when used inside a handler that is only ever called for PRIVMSG events,
// then it is safe to discard err.
// Errors are only returned to prevent the method from returning unexpected results to callers that assume it will work for all message types.
func (m *Message) Text() (string, error) {
	switch m.Command {
	case CmdQuit, CmdError:
		return m.Params.Get(1), nil
	case CmdPrivmsg, CmdNotice, CTCPAction, CmdTopic, CmdKick, CmdPart, CmdMode:
		return m.Params.Get(2), nil

	default:
		return strings.Join(m.Params, " "), fmt.Errorf("text: command %s is not supported", m.Command)
	}
}

// Target returns the intended target of a message.
// In the case of query messages, Target will equal our client's nickname.
// For channel messages, Target will usually be the name of the channel a message was sent to.
// If target is a channel,
// it may be prefixed with one or more channel membership prefixes (e.g. '@', '+' for Op, Voice)
// on servers that support the STATUSMSG response of RPL_ISUPPORT.

// Target is the target of the message, which will be the current nickname of
// the client in the case of direct messages (queries), or the channel
// name if sent to a channel, or a prefix followed by the channel name
// if sent to a specific group of users in a channel, e.g. "+#foo"
// for all users on a channel with +v or higher.
func (m *Message) Target() (string, error) {

	switch m.Command {
	case CmdPrivmsg, CmdNotice, CTCPAction, CmdInvite, CmdTopic, CmdKick, CmdPart, CmdMode:
		return m.Params.Get(1), nil
	default:
		return "", fmt.Errorf("%s: target method not supported", m.Command)
	}
}

// Chan is the channel the message was sent to. If the message was a direct
// message (query), Chan will be an empty value. If the message target
// was a group on a channel, e.g. "+#foo", then Chan will be the
// channel name with the target prefix removed ("#foo").

// Chan returns the channel a message applies to.
// In the case of query messages, Chan will return an empty string.
// If the message target was a channel name prefixed with membership prefixes ('@', '+', etc.) the prefixes will be stripped.
func (m *Message) Chan() (string, error) {

	// todo: return empty string when target wasn't a channel
	/*
		- remove status prefixes (@%+)
		- if the next character is a chanprefix (#) then it is a channel
		- else return ""
	*/
	switch m.Command {
	case CmdPrivmsg, CmdNotice, CTCPAction, CmdJoin, CmdTopic, CmdKick, CmdPart:
		return m.Params.Get(1), nil
	case CmdInvite:
		return m.Params.Get(2), nil
	default:
		return "", fmt.Errorf("%s: chan method not supported", m.Command)
	}
}
