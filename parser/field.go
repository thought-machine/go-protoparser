package parser

import (
	"strings"

	"github.com/thought-machine/go-protoparser/internal/lexer/scanner"
	"github.com/thought-machine/go-protoparser/parser/meta"
)

// FieldOption is an option for the field.
type FieldOption struct {
	OptionName string
	Constant   string
}

// Field is a normal field that is the basic element of a protocol buffer message.
type Field struct {
	IsRepeated   bool
	Type         string
	FieldName    string
	FieldNumber  string
	FieldOptions []*FieldOption

	// Comments are the optional ones placed at the beginning.
	Comments []*Comment
	// InlineComment is the optional one placed at the ending.
	InlineComment *Comment
	// Meta is the meta information.
	Meta meta.Meta
}

// SetInlineComment implements the HasInlineCommentSetter interface.
func (f *Field) SetInlineComment(comment *Comment) {
	f.InlineComment = comment
}

// Accept dispatches the call to the visitor.
func (f *Field) Accept(v Visitor) {
	if !v.VisitField(f) {
		return
	}

	for _, comment := range f.Comments {
		comment.Accept(v)
	}
	if f.InlineComment != nil {
		f.InlineComment.Accept(v)
	}
}

// ParseField parses the field.
//  field = [ "repeated" ] type fieldName "=" fieldNumber [ "[" fieldOptions "]" ] ";"
//
// See https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#field
func (p *Parser) ParseField() (*Field, error) {
	var isRepeated bool
	p.lex.NextKeyword()
	if p.lex.Token == scanner.TREPEATED {
		isRepeated = true
	} else {
		p.lex.UnNext()
	}
	startPos := p.lex.Pos

	typeValue, _, err := p.parseType()
	if err != nil {
		return nil, p.unexpected("type")
	}

	p.lex.Next()
	if p.lex.Token != scanner.TIDENT {
		return nil, p.unexpected("fieldName")
	}
	fieldName := p.lex.Text

	p.lex.Next()
	if p.lex.Token != scanner.TEQUALS {
		return nil, p.unexpected("=")
	}

	fieldNumber, err := p.parseFieldNumber()
	if err != nil {
		return nil, p.unexpected("fieldNumber")
	}

	fieldOptions, err := p.parseFieldOptionsOption()
	if err != nil {
		return nil, err
	}

	p.lex.Next()
	if p.lex.Token != scanner.TSEMICOLON {
		return nil, p.unexpected(";")
	}

	return &Field{
		IsRepeated:   isRepeated,
		Type:         typeValue,
		FieldName:    fieldName,
		FieldNumber:  fieldNumber,
		FieldOptions: fieldOptions,
		Meta:         meta.NewMeta(startPos),
	}, nil
}

// [ "[" fieldOptions "]" ]
func (p *Parser) parseFieldOptionsOption() ([]*FieldOption, error) {
	p.lex.Next()
	if p.lex.Token == scanner.TLEFTSQUARE {
		fieldOptions, err := p.parseFieldOptions()
		if err != nil {
			return nil, err
		}

		p.lex.Next()
		if p.lex.Token != scanner.TRIGHTSQUARE {
			return nil, p.unexpected("]")
		}
		return fieldOptions, nil
	}
	p.lex.UnNext()
	return nil, nil
}

// fieldOptions = fieldOption { ","  fieldOption }
// See https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#field
func (p *Parser) parseFieldOptions() ([]*FieldOption, error) {
	opt, err := p.parseFieldOption()
	if err != nil {
		return nil, err
	}

	var opts []*FieldOption
	opts = append(opts, opt)

	for {
		p.lex.Next()
		if p.lex.Token != scanner.TCOMMA {
			p.lex.UnNext()
			break
		}

		opt, err = p.parseFieldOption()
		if err != nil {
			return nil, p.unexpected("fieldOption")
		}
		opts = append(opts, opt)
	}
	return opts, nil
}

// fieldOption = optionName "=" constant
// See https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#field
func (p *Parser) parseFieldOption() (*FieldOption, error) {
	optionName, err := p.parseOptionName()
	if err != nil {
		return nil, err
	}

	p.lex.Next()
	if p.lex.Token != scanner.TEQUALS {
		return nil, p.unexpected("=")
	}

	var constant string
	switch p.lex.Peek() {
	// go-proto-validators requires this exception.
	case scanner.TLEFTCURLY:
		if !p.permissive {
			return nil, p.unexpected("constant or permissive mode")
		}

		constant, err = p.parseGoProtoValidatorFieldOptionConstant()
		if err != nil {
			return nil, err
		}
	default:
		constant, _, err = p.lex.ReadConstant()
		if err != nil {
			return nil, err
		}
	}

	return &FieldOption{
		OptionName: optionName,
		Constant:   constant,
	}, nil
}

// goProtoValidatorFieldOptionConstant = "{" ident ":" constant { , ident ":" constant } "}"
func (p *Parser) parseGoProtoValidatorFieldOptionConstant() (string, error) {
	var ret string

	p.lex.Next()
	if p.lex.Token != scanner.TLEFTCURLY {
		return "", p.unexpected("{")
	}
	ret += p.lex.Text

	for {
		p.lex.Next()
		if p.lex.Token != scanner.TIDENT {
			return "", p.unexpected("ident")
		}
		ret += p.lex.Text

		p.lex.Next()
		if p.lex.Token != scanner.TCOLON {
			return "", p.unexpected(":")
		}
		ret += p.lex.Text

		switch p.lex.Peek() {
		case scanner.TLEFTSQUARE:
			consts, constsErr := p.parseConstList()
			if constsErr != nil {
				return "", constsErr
			}
			ret += "["
			ret += strings.Join(consts, " ")
			ret += "]"
		default:
			constant, _, err := p.lex.ReadConstant()
			if err != nil {
				return "", err
			}
			ret += constant
		}

		p.lex.Next()
		switch {
		case p.lex.Token == scanner.TCOMMA:
			ret += p.lex.Text
			if p.lex.Peek() == scanner.TRIGHTCURLY {
				p.lex.Next()
				ret += p.lex.Text
				return ret, nil
			}
		case p.lex.Token == scanner.TRIGHTCURLY:
			ret += p.lex.Text
			return ret, nil
		default:
			p.lex.UnNext()
		}
	}
}

func (p *Parser) parseConstList() ([]string, error) {
	p.lex.Next()
	if p.lex.Token != scanner.TLEFTSQUARE {
		return nil, p.unexpected("[")
	}
	var consts []string
	for {
		if p.lex.Peek() == scanner.TRIGHTSQUARE {
			p.lex.Next()
			return consts, nil
		}
		p.lex.NextStrLit()
		if p.lex.Token != scanner.TSTRLIT {
			return nil, p.unexpected("String Lit")
		}
		consts = append(consts, p.lex.Text)
		if p.lex.Peek() == scanner.TCOMMA {
			p.lex.Next()
		}
	}
}

var typeConstants = map[string]struct{}{
	"double":   {},
	"float":    {},
	"int32":    {},
	"int64":    {},
	"uint32":   {},
	"uint64":   {},
	"sint32":   {},
	"sint64":   {},
	"fixed32":  {},
	"fixed64":  {},
	"sfixed32": {},
	"sfixed64": {},
	"bool":     {},
	"string":   {},
	"bytes":    {},
}

// type = "double" | "float" | "int32" | "int64" | "uint32" | "uint64"
//      | "sint32" | "sint64" | "fixed32" | "fixed64" | "sfixed32" | "sfixed64"
//      | "bool" | "string" | "bytes" | messageType | enumType
// See https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#fields
func (p *Parser) parseType() (string, scanner.Position, error) {
	p.lex.Next()
	if _, ok := typeConstants[p.lex.Text]; ok {
		return p.lex.Text, p.lex.Pos, nil
	}
	p.lex.UnNext()

	messageOrEnumType, startPos, err := p.lex.ReadMessageType()
	if err != nil {
		return "", scanner.Position{}, err
	}
	return messageOrEnumType, startPos, nil
}

// fieldNumber = intLit;
// See https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#fields
func (p *Parser) parseFieldNumber() (string, error) {
	p.lex.NextNumberLit()
	if p.lex.Token != scanner.TINTLIT {
		return "", p.unexpected("intLit")
	}
	return p.lex.Text, nil
}
