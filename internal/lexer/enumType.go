package lexer

import "github.com/thought-machine/go-protoparser/internal/lexer/scanner"

// ReadEnumType reads a messageType.
// enumType = [ "." ] { ident "." } enumName
// See https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#identifiers
func (lex *Lexer) ReadEnumType() (string, scanner.Position, error) {
	return lex.ReadMessageType()
}
