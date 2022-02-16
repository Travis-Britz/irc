package irc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

var errPingTimeout = errors.New("ping timeout")

// A Client manages a connection to an IRC server.
// It reads/writes IRC lines on the connection,
// and calls the handler for each Message it parses from the connection.
type Client struct {

	// The address ("host:port") of the IRC server. Only TLS connections are supported; use DialFn for anything else.
	// Addr is only used when DialFn is nil.
	Addr string

	// The nickname used by the Client when connecting to an IRC network (required).
	// Nicknames cannot contain spaces.
	Nickname string

	// The user name (required).
	// User cannot contain spaces.
	User string

	// The realname of the client (required).
	// Also referred to as the gecos field.
	// Realname may contain spaces
	Realname string

	// The connection password (optional: depends on the network).
	Pass string

	// DialFn is a function that accepts no parameters and returns an io.ReadWriteCloser and error.
	//
	// The returned connection can be any io.ReadWriteCloser: irc, ircs, ws, wss, a server mock, etc.
	// The only requirement is that the stream consists of CRLF-delimited IRC messages.
	//
	// When DialFn is nil, the default behavior dials Addr with tls.Dial.
	DialFn func() (io.ReadWriteCloser, error)

	// ErrorLog specifies an optional logger for errors returned from parsing and encoding messages.
	// If nil, logging is done via the log package's standard logger.
	ErrorLog *log.Logger

	// chanprefixes and statusprefixes might be passed to parsed messages in order to correctly figure out Chan() and Target()
	// todo: "#&" default and then fill from 005
	// CHANTYPES=#
	chanPrefixes string
	// will be passed to parsed messages so its Chan() method can correctly determine the channel name
	// STATUSMSG=@%+
	statusPrefixes string

	// casemap controls the comparison function used to determine if two nicknames or channels are equal after case folding.
	// todo: utf-8 default? then grab from 005 only if left blank
	// q: should this be part of the Router instead? which ones need to do channel and nickname comparisons specifically?
	// CASEMAPPING=ascii
	casemap caseMapping

	// todo: 8191 default? then update the scanner to use a buffer of this size
	// readBufferSize int

	// todo: 512 default, then pass this somehow to the Message type in WriteMessage before calling marshaltext? maybe a conditional type assertion
	// writeLineSize int

	conn    io.ReadWriteCloser
	handler Handler
	state   clientState
	wg      sync.WaitGroup

	// errC is a buffered channel of errors.
	// The channel may be nil, so senders must always have a default case if sending blocked.
	// Only the first error sent to the channel will be used.
	errC chan error
}

type caseMapping int

const (
	caseMapDefault caseMapping = iota
	caseMapAscii
	caseMapRfc1459
	caseMapRfc1459Strict
	caseMapUTF8
)

// noop performs no operation
var noop HandlerFunc = func(mw MessageWriter, m *Message) {}

// ConnectAndRun establishes a connection to the remote IRC server and sends the appropriate
// IRC protocol commands to begin the connection and capability negotiation.
//
// The Handler h is called for every incoming Message parsed from the connection.
// Handlers are called synchronously because the ordering of incoming messages matters.
//
// ConnectAndRun always returns an error, with one exception: if the client sends an IRC "QUIT"
// message followed by receiving an io.EOF from the connection, then the returned error
// will be nil.
func (c *Client) ConnectAndRun(ctx context.Context, h Handler) error {
	var (
		err     error
		cancel  context.CancelFunc
		mainctx context.Context
	)

	if c.Nickname == "" {
		panic("client nickname cannot be empty")
	}

	if c.User == "" {
		c.User = "guest"
	}

	if c.Realname == "" {
		// Realname is a required field when connecting to an IRC server,
		// but it's not important if left blank by a user of this package.
		c.Realname = "..."
	}

	if c.DialFn == nil {
		if c.Addr == "" {
			panic("ConnectAndRun: Addr cannot be empty when DialFn is nil")
		}
		c.DialFn = func() (io.ReadWriteCloser, error) {
			return tls.Dial("tcp", c.Addr, nil)
		}
	}

	// this context intentionally doesn't use ctx as a parent because we listen for ctx.Done() to trigger
	// a graceful shutdown (sending QUIT). that doesn't work if all of our goroutines have already exited.
	mainctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// initial state
	c.state = clientState{
		nick:   c.Nickname,
		user:   c.User,
		server: strings.Split(c.Addr, ":")[0],
	}

	if c.conn != nil {
		return errors.New("the client already has a connection")
	}

	if c.conn, err = c.DialFn(); err != nil {
		return err
	}
	defer func() {
		_ = c.conn.Close()
		c.conn = nil
	}()

	// trigger shutdown on the first read from the error channel
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer c.conn.Close()
		defer cancel()

		c.errC = make(chan error, 1)
		err = <-c.errC // err is used in the method return value
		c.errC = nil
	}()

	if h == nil {
		h = noop
	}

	pinger := &pingHandler{
		timeout: func() {
			c.exit(errPingTimeout)
		},
	}

	c.handler = wrap(h, ctcpHandler, pingMiddleware, pinger.pongHandler, c.state.middleware, capLSHandler)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.mainLoop(mainctx, pinger)
	}()

	// when ctx is done we try to close the connection gracefully
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		select {
		case <-mainctx.Done():
			// if mainctx is done that means an error was already read from c.errC and the client is already closing
			return
		case <-ctx.Done():
			c.WriteMessage(Quit("closing link"))
			select {
			// after sending a quit message we wait for c.errC to receive an error from the connection being closed by the server
			case <-mainctx.Done():
				// if we're still waiting, just shut down
			case <-time.After(3 * time.Second):
				c.exit(nil)
			}
		}
	}()

	c.WriteMessage(CapLS("302"))
	if c.Pass != "" {
		c.WriteMessage(Pass(c.Pass))
	}
	c.WriteMessage(Nick(c.Nickname))
	c.WriteMessage(User(c.User, c.Realname))

	c.wg.Wait()
	if err == io.EOF && c.state.status == statusDisconnecting {
		return nil
	}
	return err
}

func (c *Client) mainLoop(ctx context.Context, pinger *pingHandler) {
	// todo: move the message parsing into the reader so that can run concurrently with the main handler loop
	readLine := c.startReading(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case l, ok := <-readLine:
			if !ok {
				c.exit(errors.New("read channel closed"))
				return
			}
			m := new(Message)
			m.IncludePrefix()
			if err := m.UnmarshalText(l); err != nil {
				// A parse error might be caused by a malformed line from the remote server
				// or a bug in our message parser. Both cases are interesting but not
				// a reason to cause the client to exit.
				c.log(err)
				continue
			}
			// rfc1459: If the prefix is missing from the message, it
			// is assumed to have originated from the connection from which it was
			// received.
			if (m.Source == Prefix{}) {
				m.Source.Host = c.state.server
			}
			c.handler.SpeakIRC(c, m)
		case <-time.After(2 * time.Minute):
			// using time.After() for every line read from the connection probably isn't good,
			// but it can be cleaned up later without breaking any interfaces or behavior
			pinger.ping(ctx, c, "TIMEOUTCHECK")
		}
	}

}

func (c *Client) startReading(ctx context.Context) <-chan []byte {
	lines := make(chan []byte)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(lines)

		s := bufio.NewScanner(c.conn)
		for s.Scan() {
			l := s.Bytes()
			if len(l) == 0 {
				continue
			}
			select {
			case <-ctx.Done():
				// the main loop could have returned before the reader, so we need another way out so that lines <- l doesn't block.
				// if the loop is sitting inside s.Scan() we won't actually be able to read from ctx.Done() until another
				// line is read from the connection. the ping timeout will usually trigger this eventually from idle connections,
				// (and if the main loop already exited then it will always select the ctx.Done() case)
				// but to exit in a timely manner the connection will need to be closed to break s.Scan().
				return
			case lines <- l:
			}
		}
		err := s.Err()
		// scanner.Err() returns nil when the reader error was EOF, but the IRC client
		// wants to know when the error is EOF in order to determine if the
		// connection was terminated gracefully.
		if err == nil {
			c.exit(io.EOF)
		} else {
			c.exit(err)
		}
	}()
	return lines
}

// exit requests the client to exit and return with err. Only the first such error
// is returned; any successive calls to exit will drop the error, such as if
// there were remaining writes that also failed with errors.
func (c *Client) exit(err error) {
	select {
	case c.errC <- err:
	default:
	}
}

// WriteMessage implements irc.MessageWriter.
// It writes m to the client's connection.
// Marshaling errors will be reported to the client's logger.
// Write errors will cause the client's run method to return with the first error.
func (c *Client) WriteMessage(m encoding.TextMarshaler) {
	// WriteMessage does not return any errors itself because IRC itself does not provide any guarantees about message delivery.
	// Even if bytes are successfully written to a TCP stream, that does not guarantee message delivery to the intended recipient(s).
	//
	var (
		err error
		b   []byte
	)

	if c.conn == nil {
		c.log(fmt.Errorf("WriteMessage: conn cannot be nil; m: %#v", m))
		return
	}

	if msg, ok := m.(*Message); ok && !msg.includePrefix {
		// set the message prefix to what the client thinks it is currently
		// so that marshaltext can correctly return warnings when lines are likely to be truncated
		msg.Source = c.prefix()
	}

	b, err = m.MarshalText()
	if err != nil {
		c.log(fmt.Errorf("marshal text: %w; message: %#v", err, m))
		return
	}
	if !bytes.HasSuffix(b, []byte("\r\n")) {
		b = append(b, []byte("\r\n")...)
	}

	// this might not be the cleanest way to intercept outgoing quit commands,
	// but it works for now and lets us rewrite ConnectAndRun's error to nil
	// when the exit was intentional
	if bytes.HasPrefix(b, []byte("QUIT")) {
		c.state.status = statusDisconnecting
	}

	if _, err = c.conn.Write(b); err != nil {
		c.exit(err)
	}
}

// log reports errors which are noteworthy but not a reason for the client to exit.
func (c *Client) log(e error) {
	if c.ErrorLog == nil {
		log.Println(e)
		return
	}
	c.ErrorLog.Println(e)
}

// clientState groups and manages access to a minimal set of
// state around each new connection to the IRC server.
type clientState struct {

	// the client's current nickname, used for calculating max outgoing message length and for
	// matching events that originated from our client.
	nick string

	// the client's user as seen by the server, used for calculating max outgoing message length.
	// this may differ from the name defined in Client.cfg on servers which use an
	// ident service to verify the user name. Such servers typically prefix
	// the user name with a tilde (~) to indicate the ident was not
	// validated against an identd server.
	user string

	// the client's host as seen by the server, used for calculating max outgoing message length.
	host string

	// the server the client is connected to, used as the message source when incoming messages didn't contain a prefix.
	server string

	// status contains the client's connection state: disconnected, connected, etc.
	// not all states are implemented.
	// only the "disconnecting" state is used to rewrite io.EOF errors to nil when the disconnect was intentional
	status clientStatus
}

// Nick returns the client's current nickname according to the client's internal state tracking.
// This is used by some route matchers to determine when a message originated from or targeted our client.
func (c *Client) Nick() Nickname {
	return Nickname(c.state.nick)
}

// prefix returns the estimated prefix based on internal state tracking,
// used by Message to calculate the actual limit of outgoing messages.
func (c *Client) prefix() Prefix {
	return Prefix{
		Nick: Nickname(c.state.nick),
		Host: c.state.host,
		User: c.state.user,
	}
}

var fullAddress = regexp.MustCompile("^([^!@]+)!(.+?)@(.+)?$")

// stateMiddleware intercepts various events to keep the client state up to date.
func (s *clientState) middleware(next Handler) Handler {
	return HandlerFunc(func(mw MessageWriter, m *Message) {
		switch m.Command {

		// By saving our host (as seen by the server) we can more accurately calculate the maximum length
		// of any message we can send, because 512-byte line length limit defined by the IRC protocol
		// will include our nickname and host in each message when they are received by others.
		//
		// Format: "Welcome to the Internet Relay Network <nick>!<user>@<host>"
		case RplWelcome:
			fields := strings.Fields(m.Params.Get(2))
			if len(fields) == 0 {
				fields = []string{""}
			}
			// The last field can be <nick> or <nick>!<user>@<host>, but the format of RPL_WELCOME varies so widely that
			// accepting anything other than nick!user@host might break our nick state tracking.
			// For example, twitch.tv servers ignore the spec completely and include neither
			// the network name nor our nickname.
			if parts := fullAddress.FindStringSubmatch(fields[len(fields)-1]); parts != nil {
				s.nick = parts[1]
				s.user = parts[2]
				s.host = parts[3]
			}
		case RplMyInfo:
			// Even though param 2 should contain the server host, checking for more than 2 params is a smoke test
			// to determine if the line is likely to follow protocol. If not, we'll fall back and hope
			// the message prefix contains the server host. Twitch.tv notably breaks protocol here
			// by sending only 2 params, with the second being only a single hyphen (-).
			// Even though the twitch case doesn't technically matter because their
			// server and host names are static, it annoyed me that the wrong
			// info would be contained in the client state.
			if len(m.Params) > 2 {
				s.server = m.Params.Get(2)
			} else {
				s.server = m.Source.Host
			}
		case RplHostHidden:
			// "<target> <host> :is now your displayed host"
			// Some servers implement numeric 396 to indicate when our displayed host is changed,
			// e.g. hidden or unhidden with user mode +x/-x. We listen for this by default to
			// improve our calculations for the maximum message length we can send.
			if len(m.Params) > 1 {
				s.host = m.Params.Get(2)
			}
		case CmdNick:
			if m.Source.Nick.Is(s.nick) {
				s.nick = m.Params.Get(1)
			}
		}

		next.SpeakIRC(mw, m)
	})
}

type clientStatus int

func (s clientStatus) String() string {
	switch s {
	case statusDisconnected:
		return "disconnected"
	case statusConnecting:
		return "connecting"
	case statusConnected:
		return "connected"
	case statusDisconnecting:
		return "disconnecting"
	default:
		return "unknown"
	}
}

const (
	statusDisconnected clientStatus = iota
	statusConnecting
	statusConnected
	statusDisconnecting
)
