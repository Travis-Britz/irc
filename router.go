package irc

import (
	"fmt"
	"regexp"
	"strings"
)

// Router provides a Handler which can match incoming messages against a slice of route handlers.
// Matching is based on message attributes such as the command (verb), source, message contents, and more.
//
// Routes are currently tested in the order they were added, and only the first matching route's handler
// will be called. However, this behavior may change in the future to allow for more efficient route
// matching. Therefore, care should be taken to avoid adding multiple routes which may trigger
// on the same input message.
type Router struct {

	// routes to be matched, in order.
	routes []*route

	// Slice of middleware to be called, regardless of whether a match was found.
	middlewares []middleware

	// chanmodes and nickprefixes are used to split MODE messages into multiple events
	// CHANMODES=A,B,C,D[,X,Y...]
	// CHANMODES=beIqa,kLf,lH,psmntirzMQNRTOVKDdGPZSCc
	// combine with the PREFIX modes, which are always type B
	chanModes string
	// todo: "@+" default and fill from 005
	// unless I don't actually use this anywhere?
	// PREFIX=(ohv)@%+
	nickPrefixes string
}

var defaultChanModes = chanModes{
	A: "beI",
	B: "k",
	C: "l",
	D: "psitnm",
}

type chanModes struct {
	A string
	B string
	C string
	D string
}

func handleMode(next Handler) Handler {

	// if 005, grab the PREFIX and CHANMODES and update our types
	// if MODE:
	// loop through + mode changes, for each character check if it needs a param and consume a parameter
	// pass a new mode message to the next handler
	// loop through - mode changes, for each character check if it needs a param and consume a parameter
	// pass a new mode message to the next handler
	// return without calling the next handler with the original message
	return next
}

func (r *Router) OnOp(h HandlerFunc) *route {
	// todo: match channel
	// OnOp depends on mode commands being split into multiple events before hitting the router
	return r.HandleFunc(CmdMode, h).MatchFunc(func(m *Message) bool {
		return strings.HasPrefix(m.Params.Get(2), "+o ")
	})
}

// Handle appends h to the list of handlers for cmd.
func (r *Router) Handle(cmd Command, h Handler) *route {
	rt := &route{
		h:        h,
		matchers: []matcher{&commandMatch{cmd}},
	}
	r.routes = append(r.routes, rt)
	return rt
}

// HandleFunc appends f to the list of handlers for cmd.
func (r *Router) HandleFunc(cmd Command, f HandlerFunc) *route {
	return r.Handle(cmd, f)
}

// SpeakIRC implements Handler
func (r *Router) SpeakIRC(mw MessageWriter, m *Message) {

	for _, rt := range r.routes {
		if rt.matches(m) {
			wrap(rt.h, r.middlewares...).SpeakIRC(mw, m)
			return
		}
	}
	// global middlewares need to run even if there was no matching route
	// since there's no route handler, we wrap the no-op handler
	wrap(noop, r.middlewares...).SpeakIRC(mw, m)
}

// Use appends global middleware to the router.
// Middleware are functions which accept a handler and return a handler.
//
// Global middleware are run against every incoming line,
// even if there were no matching routes for the message.
//
// Middleware can do many things:
//
//  - Mutate incoming messages before passing them to the next Handler
//  - Decorate the MessageWriter with additional functionality before passing it to the next Handler
//  - Write messages to the MessageWriter
//  - Prevent additional processing by not calling the next Handler
//
// These are very powerful abilities, but it is very easy to use them improperly.
//
// Middleware will execute in the order they were attached.
func (r *Router) Use(middlewares ...middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// Use wraps the route handler with middlewares.
// The given middlewares will execute in the order listed.
// The middleware only execute if the route matched.
//
// Route-specific middleware are ideal for generic functionality that might be shared among many routes, such as:
//
//  - rate limiting command usage
//  - checking if your bot is an Op in a channel before performing an action like kicking/banning
//  - checking if some service you depend on has an active connection (and sending an error when it's offline)
//  - checking if the nick that sent the message is authorized to trigger a handler
//  - stripping message formatting such as special control characters for colors, bold, italics, etc. before passing it to the handler
//
// Use panics if the route handler is nil.
func (r *route) Use(middlewares ...middleware) *route {
	if r.h == nil {
		panic("nil handler: the route handler must be defined before wrapping the handler with middleware")
	}
	r.h = wrap(r.h, middlewares...)
	return r
}

// OnConnect attaches a handler which is called upon successful connection to an IRC server, after
// capability negotiation is complete (on servers which support capability negotiation).
// More specifically, it is triggered by numeric 001 (RPL_WELCOME).
func (r *Router) OnConnect(h HandlerFunc) *route {
	return r.Handle(RplWelcome, h)
}

// OnText attaches a handler for PRIVMSG events that match text. text is a wildcard string:
//
//  * matches any text
//  & matches any word
//  ? matches a single character
//  text matches if exact match
//  text* matches if text starts with word
//  *text matches if text ends with word
//  *text* matches if text is anywhere
func (r *Router) OnText(wildtext string, h HandlerFunc) *route {
	return r.HandleFunc(CmdPrivmsg, h).wildtext(wildtext)
}

// OnTextRE attaches the handler h for PRIVMSG events that match the Go regular expression expr.
func (r *Router) OnTextRE(expr string, h HandlerFunc) *route {
	return r.HandleFunc(CmdPrivmsg, h).textRE(expr)
}

// OnNotice is triggered when a NOTICE is received from a client on the server, following the
// same format as OnText. For server notices, use MatchServer.
func (r *Router) OnNotice(wildtext string, h HandlerFunc) *route {
	return r.HandleFunc(CmdNotice, h).
		wildtext(wildtext).
		MatchFunc(func(m *Message) bool {
			return !m.Source.IsServer()
		})
}

// OnAction attaches a handler for PRIVMSG that matches CTCP ACTION, and follows the same
// format as OnText.
func (r *Router) OnAction(wildtext string, h HandlerFunc) *route {
	return r.HandleFunc(CTCPAction, h).wildtext(wildtext)
}

// OnJoin attaches a handler for JOIN events.
func (r *Router) OnJoin(h HandlerFunc) *route {
	return r.Handle(CmdJoin, h)
}

// OnPart is triggered when a client departs a channel we are on.
func (r *Router) OnPart(h HandlerFunc) *route {
	return r.Handle(CmdPart, h)
}

// OnQuit is triggered when a client which shares a channel with us disconnects from the server.
func (r *Router) OnQuit(h HandlerFunc) *route {
	return r.Handle(CmdQuit, h)
}

// OnError is triggered when the server sends an ERROR message, usually on disconnect.
func (r *Router) OnError(h HandlerFunc) *route {
	return r.Handle(CmdError, h)
}

// OnNick attaches a handler when a user's nickname changes.
func (r *Router) OnNick(h func(nick Nickname, newnick Nickname)) *route {
	adapter := func(mw MessageWriter, m *Message) {
		h(m.Source.Nick, Nickname(m.Params.Get(1)))
	}
	return r.HandleFunc(CmdNick, adapter)
}

// OnCTCP attaches a route handler that matches against a CTCP message of type subcommand.
func (r *Router) OnCTCP(subcommand string, h HandlerFunc) *route {
	return r.Handle(NewCTCPCmd(subcommand), h)
}

// OnCTCPReply attaches a route handler that matches against a CTCP Reply of type subcommand.
func (r *Router) OnCTCPReply(subcommand string, h HandlerFunc) *route {
	return r.Handle(NewCTCPReplyCmd(subcommand), h)
}

// NewCTCPCmd returns a Command which will match the internal representation of a CTCP-encoded
// PRIVMSG, for mapping CTCP messages to handlers.
//
// The returned Command is *not* a valid IRC commnad. For sending CTCP-formatted messages, see func CTCP.
// todo: rename this and newctcpreplycmd so it doesn't get confused with the Message builder?
func NewCTCPCmd(subcommand string) Command {
	return Command(fmt.Sprintf("_CTCP_QUERY_%s", strings.ToUpper(subcommand)))
}

// NewCTCPReplyCmd returns a Command which will match the internal representation of a CTCP-encoded
// NOTICE, for mapping CTCP replies to handlers.
//
// The returned Command is *not* a valid IRC command. For sending CTCP-formatted replies, see func CTCPReply.
func NewCTCPReplyCmd(subcommand string) Command {
	return Command(fmt.Sprintf("_CTCP_REPLY_%s", strings.ToUpper(subcommand)))
}

type route struct {
	h        Handler
	matchers []matcher
}

func (r *route) matches(m *Message) bool {
	for _, rm := range r.matchers {
		if !rm.matches(m) {
			return false
		}
	}
	return true
}

// A matcher is attached to a route and determines whether a given Message satisfies some condition.
type matcher interface {
	matches(*Message) bool
}

// wildtext converts a wildcard match string to a regex match string.
//
// Rules
//
// * matches any text
// & matches any word (delimited by ascii space)
// ? matches a single character
// text matches if exact match
// text* matches if text starts with word
// *text matches if text ends with word
// *text* matches if text is anywhere
func (r *route) wildtext(s string) *route {

	re := regexp.MustCompile("\\*|\\?|[^*?]+")
	expr := re.ReplaceAllStringFunc(s, func(s string) string {
		switch s {
		case "*":
			return ".*"
		case "?":
			return "."
		}
		return regexp.QuoteMeta(s)
	})

	fields := strings.Split(expr, " ")
	for i, f := range fields {
		if f == "&" {
			fields[i] = "\\S+"
		}
	}

	expr = strings.Join(fields, " ")

	return r.textRE("^" + expr + "$")
}

func (r *route) matchtext(s string) *route {
	return r.wildtext(s)
}

// textRE appends the regular expression expr to the route's matchers.
func (r *route) textRE(expr string) *route {
	r.matchers = append(r.matchers, &regexMatch{regexp.MustCompile(expr)})
	return r
}

type nickTracker interface {
	Nick() Nickname
}

// isQuery limits the route to match only against query messages
func (r *route) isQuery() *route {

	var nt nickTracker = nil // todo: figure out how to cleanly pass in a reference to the client

	if nt == nil {
		panic("isQuery: the router's nick tracker cannot be nil when using matchers that need the client's current nickname")
	}
	r.MatchFunc(func(m *Message) bool {
		targ, err := m.Target()
		if err != nil {
			return false
		}
		return nt.Nick().Is(targ)
	})
	return r
}

func (r *route) channel(ch string) *route {
	// not exported yet because I'm not sure how to deal with events other than privmsg/notice
	r.matchers = append(r.matchers, &channelMatch{ch})
	return r
}
func (r *route) MatchFunc(f matcherFunc) *route {
	return r.Matcher(f)
}

func (r *route) MatchServer() *route {
	return r.MatchFunc(func(m *Message) bool {
		return m.Source.IsServer()
	})
}

func (r *route) Matcher(m matcher) *route {
	r.matchers = append(r.matchers, m)
	return r
}

func (r *route) MatchChan(ch string) *route {
	return r.channel(ch)
}

type matchAny struct {
	matchers []matcher
}

func (ma *matchAny) matches(m *Message) bool {
	for _, rm := range ma.matchers {
		if rm.matches(m) {
			return true
		}
	}
	return false
}

// MatchClient matches the source of a message against the client's current nickname.
// todo: rename? MatchMe(), MatchSource(), MatchNick() - MatchNick() might be the most generic, especially with the EventNick interface?
func (r *route) MatchClient(client nickTracker) *route {
	return r.MatchFunc(func(m *Message) bool {
		switch m.Command {
		case CmdKick:
			return client.Nick().Is(m.Params.Get(2))
		default:
			return m.Source.Nick.Is(client.Nick().String())
		}
	})
}

type commandMatch struct {
	cmd Command
}

type matcherFunc func(m *Message) bool

func (f matcherFunc) matches(m *Message) bool {
	return f(m)
}

func (cm commandMatch) matches(m *Message) bool {
	return m.Command.is(cm.cmd)
}

type regexMatch struct {
	re *regexp.Regexp
}

func (rm regexMatch) matches(m *Message) bool {
	text, err := m.Text()
	if err != nil {
		return false
	}
	return rm.re.MatchString(text)
}

type channelMatch struct {
	channel string
}

func (cm channelMatch) matches(m *Message) bool {
	ch, err := m.Chan()
	if err != nil {
		return false
	}
	return strings.EqualFold(cm.channel, ch)
}
