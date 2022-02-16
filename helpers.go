package irc

// EqualFold tests whether two strings are equal according to mapping.
// func EqualFold(s1, s2 string, mapping caseMapping) bool {
//
// }

// MaskToRegex converts an IRC wildcard expression to its equivalent regex
// '?' matches one and only one character, and '*' matches any number of characters.
// These characters can be escaped with the '\' character.
// https://modern.ircdocs.horse/#wildcard-expressions
// func MaskToRegex(mask string) string {

// }

// Mask converts a full address into an address mask of type maskType.
// func Mask(fulladdress string, maskType int) string {
//
// }
//
// IsWM compares a wildcard string with an input string and determines whether text matches wildText.
// func IsWM(wildText string, text string) bool {
//
// }
//
// StripColors removes IRC color codes from text.
// func StripColors(text string) string {
//
// }
//
// StripFormatting removes IRC formatting control characters from text.
// func StripFormatting(text string) string {
//
// }
//
// func Colorize(text string, fg int, bg int) string {
//
// }

// Decode decodes a line of IRC text into a Message struct. line must not end with line endings \r\n.
// func Decode(line []byte) (*Message, error) {
// 	m := new(Message)
// 	err := m.UnmarshalText(line)
// 	return m, err
// }

// Encode encodes a message to be sent on an IRC connection.
// func Encode(command string, params ...string) ([]byte,error) {
// 	m := NewMessage(Command(command), params...)
// 	return m.MarshalText()
// }
