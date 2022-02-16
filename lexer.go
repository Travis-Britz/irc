// This lexer follows the method described in the video:
// Lexical Scanning in Go - Rob Pike
// https://www.youtube.com/watch?v=HxaD_trXwRE

package irc

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	delimParam    = ' ' // the delimiter token for parameters
	delimTag      = ';' // the delimiter token for message tags
	delimTagValue = '=' // the delimiter token for message tag values
	startTags     = '@' // the delimiter for beginning tags
	startPrefix   = ':' // the delimiter for the prefix
	startTrailing = ':' // the delimiter for the trailing param
)

// item represents a token returned from the scanner.
type item struct {
	typ itemType // Type, such as itemTarget
	val string   // the value of the lexed token
}

func (it itemType) String() string {
	switch it {
	case itemCommand:
		return "Command"
	case itemPrefix:
		return "Source"
	case itemNickname:
		return "Nickname"
	case itemUser:
		return "User"
	case itemHost:
		return "Host"
	case itemTagKey:
		return "TagKey"
	case itemTagValue:
		return "TagValue"
	case itemParam:
		return "Param"
	case itemError:
		return "Error"
	default:
		return ""
	}
}
func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	}

	return fmt.Sprintf("%s: %q", i.typ, i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred;
	// value is text of error. Rob Pike had a specific error item so I'll assume it comes in handy later.
	itemTagKey   // IRCv3 message tag key
	itemTagValue // message tag value
	itemNickname
	itemUser
	itemHost
	itemPrefix  // the prefix portion of a message, e.g. ":tmi.switch.tv"
	itemCommand // the command or numeric, e.g. "PRIVMSG" or "001"
	itemParam   // a command parameter, e.g. the target and text of a PRIVMSG
	itemEOF     // end of message
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name  string    // used only for error reports.
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of the last rune read
	items chan item // channel of scanned item
}

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) ignoreRun(run string) {
	l.acceptRun(run)
	l.ignore()
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r

}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}

	// todo: figure out what this todo was. it was empty last time I was here, which is ominous. what did I need to remember???
	go l.run() // Concurrently run state machine
	return l
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	return <-l.items
}

func lexStart(l *lexer) stateFn {
	if l.peek() == startTags {
		return lexTagsStart
	}
	if l.peek() == startPrefix {
		return lexPrefixStart
	}
	return lexCommand
}

func lexTagsStart(l *lexer) stateFn {
	l.pos++    // we know the delimiters are single-byte because the protocol is from the days of ascii
	l.ignore() // drop the @
	return lexTagKey
}

// lexTagKey lexes an IRCv3 message tag.
//
// The caller is responsible for removing empty keys.
// itemTagKey is always followed by itemTagValue.
func lexTagKey(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == delimTagValue:
			l.backup()
			l.emit(itemTagKey)
			return lexTagDelimiter
		case r == delimTag || r == delimParam:
			l.backup()
			l.emit(itemTagKey)
			return lexTagValue
		case r == eof:
			return l.errorf("unexpected end of input while reading tag name")
		case invalidTagNameChar(r):
			return l.errorf("invalid character %q found while reading tag name", r)
		}
	}
}
func lexTagDelimiter(l *lexer) stateFn {
	l.pos++
	l.ignore()
	return lexTagValue
}
func invalidTagNameChar(r rune) bool {
	// <key_name>      ::= <non-empty sequence of ascii letters, digits, hyphens ('-')>
	// https://ircv3.net/specs/extensions/message-tags.html
	switch r {
	// for now, include <client prefix>, <vendor>, and the / as part of the tag name
	// todo: split client prefix and vendor into separate lexemes?
	case '+', '/', '.':
		return false
	default:
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-')
	}
}

func lexTagValue(l *lexer) stateFn {
	// <escaped_value> ::= <sequence of zero or more utf8 characters except NUL, CR, LF, semicolon (`;`) and SPACE>
	// https://ircv3.net/specs/extensions/message-tags.html
	for {
		switch r := l.next(); {
		case r == delimTag:
			l.backup()
			l.emit(itemTagValue)
			return lexTagEnd
		case r == delimParam:
			l.backup()
			l.emit(itemTagValue)
			l.ignoreRun(" ")
			if l.peek() == startPrefix {
				return lexPrefixStart
			}
			if l.peek() == eof {
				return l.errorf("unexpected end of input after message tags")
			}
			return lexCommand
		case r == eof:
			return l.errorf("unexpected end of input while reading tag value")
		}
	}
}

// tagDelimiter ;
// tagValueDelimiter =

// lexTagEnd scans a tag end. The semicolon is known to be present.
func lexTagEnd(l *lexer) stateFn {
	l.pos++
	l.ignore()
	if l.peek() == delimParam {
		l.ignoreRun(" ")
		if l.peek() == startPrefix {
			return lexPrefixStart
		}
		if l.peek() == eof {
			return l.errorf("unexpected end of input after message tags")
		}
		return lexCommand
	}
	return lexTagKey
}

// lexPrefixStart scans a prefix delimiter, which is known to be present.
func lexPrefixStart(l *lexer) stateFn {
	l.pos++
	l.ignore()
	return lexNickname
}

// lexNickname scans the nickname portion of a message prefix
func lexNickname(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == delimParam:
			l.backup()
			l.emit(itemNickname)
			l.ignoreRun(" ")
			if l.peek() == eof {
				return l.errorf("unexpected end of input; expected command")
			}
			return lexCommand
		case r == '.':
			// '.' is invalid inside a nickname, which means the prefix was in the :host form
			return lexHost
		case r == '!':
			l.backup()
			l.emit(itemNickname)
			return lexUserStart
		case r == eof:
			return l.errorf("unexpected end of input")
		}
	}
}

// lexUserStart scans the user delimiter from a message prefix, which is known to be present
func lexUserStart(l *lexer) stateFn {
	l.pos++
	l.ignore()
	return lexUser
}

// lexUser scans the user portion of a message prefix. the prefix is known to be in the fulladdress form
func lexUser(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '@':
			l.backup()
			l.emit(itemUser)
			return lexHostStart
		case r == delimParam:
			return l.errorf("expected host, found end of prefix")
		case r == eof:
			return l.errorf("unexpected end of input")
		}
	}
}

// lexHostStart scans the host delimiter inside a fulladdress prefix. the @ is known to be present
func lexHostStart(l *lexer) stateFn {
	l.pos++
	l.ignore()
	return lexHost
}

// lexHost scans the host of a message prefix.
func lexHost(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == delimParam:
			l.backup()
			l.emit(itemHost)
			l.ignoreRun(" ")
			return lexCommand
		case r == eof:
			return l.errorf("expected command, found end of input")
		}
	}
}

func lexCommand(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == delimParam:
			l.backup()
			if len(l.input[l.start:l.pos]) == 0 {
				return l.errorf("unexpected end of command; command is empty")
			}
			l.emit(itemCommand)
			l.ignoreRun(" ")
			return lexParam
		case r == eof:
			if len(l.input[l.start:l.pos]) == 0 {
				return l.errorf("unexpected eof; command is empty")
			}
			l.emit(itemCommand)
			l.emit(itemEOF)
			return nil
		}
	}
}

func lexParam(l *lexer) stateFn {

	if l.peek() == startPrefix {
		return lexTrailingPrefix
	}

	for {
		switch r := l.next(); {
		case r == delimParam:
			l.backup()
			l.emit(itemParam)
			l.ignoreRun(" ")
			return lexParam
		// todo: clean up comment wording, make more concise
		// [...] any sane caller will treat an explicit empty string the same as an implied (omitted) empty string when reading a parameter [...]
		// We technically behave differently by emitting an empty param when the command or final param
		// ended with a trailing delimiter and then eof. However, this should be fine for two reasons.
		// First, arguments to IRC commands rely on their position, so a sane parser trying to read
		// from a position past what was included should always return a default (empty string)
		// when the value was missing. In other words, it should not matter whether the parameter was explicitly set to an empty string because reading from an omitted parameter still
		// returns an implied empty string.
		// Second, a trailing delimiter may indicate the intended value was actually an empty
		// string, and the encoding function on the remote end simply did not trim the
		// final delimiter before transmission.
		case r == eof:
			l.emit(itemParam)
			l.emit(itemEOF)
			return nil
		}
	}
}

func lexTrailingPrefix(l *lexer) stateFn {
	l.pos++
	l.ignore()
	return lexTrailingParam
}

func lexTrailingParam(l *lexer) stateFn {
	l.pos += len(l.input[l.pos:])
	l.emit(itemParam)
	l.emit(itemEOF)
	return nil
}
