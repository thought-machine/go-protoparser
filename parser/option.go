package parser

import (
	"github.com/thought-machine/go-protoparser/internal/lexer/scanner"
	"github.com/thought-machine/go-protoparser/parser/meta"
)

// Option can be used in proto files, messages, enums and services.
type Option struct {
	OptionName string
	Constant   string
	Endpoint   *CloudEndpoint

	// Comments are the optional ones placed at the beginning.
	Comments []*Comment
	// InlineComment is the optional one placed at the ending.
	InlineComment *Comment
	// Meta is the meta information.
	Meta meta.Meta
}

// SetInlineComment implements the HasInlineCommentSetter interface.
func (o *Option) SetInlineComment(comment *Comment) {
	o.InlineComment = comment
}

// Accept dispatches the call to the visitor.
func (o *Option) Accept(v Visitor) {
	if !v.VisitOption(o) {
		return
	}

	for _, comment := range o.Comments {
		comment.Accept(v)
	}
	if o.InlineComment != nil {
		o.InlineComment.Accept(v)
	}
}

// CloudEndpoint struct
type CloudEndpoint struct {
	Fields            []*EndpointFieldOption
	AdditionalBinding []*AdditionalBinding
}

//EndpointFieldOption struct
type EndpointFieldOption struct {
	OptionName string
	Constant   string
	Meta       meta.Meta
}

// ParseOption parses the option.
//  option = "option" optionName  "=" constant ";"
//
// See https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#option
func (p *Parser) ParseOption() (*Option, error) {
	p.lex.NextKeyword()
	if p.lex.Token != scanner.TOPTION {
		return nil, p.unexpected("option")
	}
	startPos := p.lex.Pos

	optionName, err := p.parseOptionName()
	if err != nil {
		return nil, err
	}

	p.lex.Next()
	if p.lex.Token != scanner.TEQUALS {
		return nil, p.unexpected("=")
	}

	var constant string
	var endpoint *CloudEndpoint
	switch p.lex.Peek() {
	// Cloud Endpoints requires this exception.
	case scanner.TLEFTCURLY:
		if !p.permissive {
			return nil, p.unexpected("constant or permissive mode")
		}

		endpoint, err = p.parseCloudEndpointsOptionConstant()
		if err != nil {
			return nil, err
		}
	default:
		constant, _, err = p.lex.ReadConstant(p.permissive)
		if err != nil {
			return nil, err
		}
	}

	p.lex.Next()
	if p.lex.Token != scanner.TSEMICOLON {
		return nil, p.unexpected(";")
	}

	return &Option{
		OptionName: optionName,
		Constant:   constant,
		Endpoint:   endpoint,
		Meta:       meta.NewMetaWithLastPos(startPos, p.lex.Pos),
	}, nil
}

// cloudEndpointsOptionConstant = "{" ident ":" constant { [","] ident ":" constant } "}"
//
// See https://cloud.google.com/endpoints/docs/grpc-service-config/reference/rpc/google.api
func (p *Parser) parseCloudEndpointsOptionConstant() (*CloudEndpoint, error) {

	p.lex.Next()
	if p.lex.Token != scanner.TLEFTCURLY {
		return nil, p.unexpected("{")
	}
	var endpointFields []*EndpointFieldOption
	var addBinding []*AdditionalBinding

	for {
		p.lex.NextKeyword()
		if p.lex.Token == scanner.TADDITIONAL {
			var addErr error
			addBinding, addErr = p.ParseAdditionalBindings()
			if addErr != nil {
				return nil, addErr
			}
		} else {
			p.lex.UnNext()
			p.lex.Next()
			pos := p.lex.Pos
			if p.lex.Token != scanner.TIDENT {
				return nil, p.unexpected("ident")
			}
			ident := p.lex.Text

			p.lex.Next()

			if p.lex.Token != scanner.TCOLON {
				return nil, p.unexpected(":")
			}
			
			constant := ""
			if p.lex.Peek() == scanner.TLEFTCURLY {
				endpoint, err := p.parseCloudEndpointsOptionConstant()
				if err != nil {
					return nil, err
				}
				constant = OptionConstantToString(endpoint)
			} else {
				cons, _, err := p.lex.ReadConstant(p.permissive)
				if err != nil {
					return nil, err
				}
				constant = cons
			}

			endpointFields = append(endpointFields, &EndpointFieldOption{
				OptionName: ident,
				Constant:   constant,
				Meta:       meta.NewMeta(pos),
			})
		}

		p.lex.Next()
		switch {
		case p.lex.Token == scanner.TCOMMA:
		case p.lex.Token == scanner.TRIGHTCURLY:
			return &CloudEndpoint{
				Fields:            endpointFields,
				AdditionalBinding: addBinding,
			}, nil
		default:
			p.lex.UnNext()
		}
	}
}

func OptionConstantToString(endpoint *CloudEndpoint) (string) {
 	result := "{"
	for _, field := range endpoint.Fields {
		result += field.OptionName + ":" + field.Constant + ","
	}
	result += "}"
	return result
}

// optionName = ( ident | "(" fullIdent ")" ) { "." ident }
func (p *Parser) parseOptionName() (string, error) {
	var optionName string

	p.lex.Next()
	switch p.lex.Token {
	case scanner.TIDENT:
		optionName = p.lex.Text
	case scanner.TLEFTPAREN:
		optionName = p.lex.Text
		fullIdent, _, err := p.lex.ReadFullIdent()
		if err != nil {
			return "", err
		}
		optionName += fullIdent

		p.lex.Next()
		if p.lex.Token != scanner.TRIGHTPAREN {
			return "", p.unexpected(")")
		}
		optionName += p.lex.Text
	default:
		return "", p.unexpected("ident or left paren")
	}

	for {
		p.lex.Next()
		if p.lex.Token != scanner.TDOT {
			p.lex.UnNext()
			break
		}
		optionName += p.lex.Text

		p.lex.Next()
		if p.lex.Token != scanner.TIDENT {
			return "", p.unexpected("ident")
		}
		optionName += p.lex.Text
	}
	return optionName, nil
}

// AdditionalBinding store additional binding field details
type AdditionalBinding struct {
	name   string
	values []string
}

// ParseAdditionalBindings parses a block describing additional bindings
func (p *Parser) ParseAdditionalBindings() ([]*AdditionalBinding, error) {
	p.lex.Next()
	if p.lex.Token != scanner.TLEFTCURLY {
		return nil, p.unexpected("{")
	}

	var bindings []*AdditionalBinding

	for {
		ident, _, identErr := p.lex.ReadFullIdent()

		if identErr != nil {
			return nil, identErr
		}

		p.lex.Next()

		if p.lex.Token != scanner.TCOLON {
			return nil, p.unexpected(":")
		}

		var values []string

		constVal, _, constErr := p.lex.ReadConstant(p.permissive)

		if constErr != nil {
			return nil, constErr
		}

		values = append(values, constVal)

		for {
			p.lex.NextLit()
			if p.lex.Token != scanner.TSTRLIT {
				p.lex.UnNext()
				break
			}
			values = append(values, p.lex.Text)
		}

		bindings = append(bindings, &AdditionalBinding{
			name:   ident,
			values: values,
		})

		if p.lex.Peek() == scanner.TRIGHTCURLY {
			p.lex.Next()
			break
		}
	}
	return nil, nil
}
