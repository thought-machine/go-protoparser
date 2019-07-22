package unordered

import (
	"fmt"

	"github.com/thought-machine/go-protoparser/parser"
	"github.com/thought-machine/go-protoparser/parser/meta"
)

// ServiceBody is unordered in nature, but each slice field preserves the original order.
type ServiceBody struct {
	Options []*parser.Option
	RPCs    []*parser.RPC
}

// Service consists of RPCs.
type Service struct {
	ServiceName string
	ServiceBody *ServiceBody

	// Comments are the optional ones placed at the beginning.
	Comments []*parser.Comment
	// InlineComment is the optional one placed at the ending.
	InlineComment *parser.Comment
	// InlineCommentBehindLeftCurly is the optional one placed behind a left curly.
	InlineCommentBehindLeftCurly *parser.Comment
	// Meta is the meta information.
	Meta meta.Meta
}

// InterpretService interprets *parser.Service to *Service.
func InterpretService(src *parser.Service) (*Service, error) {
	if src == nil {
		return nil, nil
	}

	serviceBody, err := interpretServiceBody(src.ServiceBody)
	if err != nil {
		return nil, err
	}
	return &Service{
		ServiceName:                  src.ServiceName,
		ServiceBody:                  serviceBody,
		Comments:                     src.Comments,
		InlineComment:                src.InlineComment,
		InlineCommentBehindLeftCurly: src.InlineCommentBehindLeftCurly,
		Meta:                         src.Meta,
	}, nil
}

func interpretServiceBody(src []parser.Visitee) (
	*ServiceBody,
	error,
) {
	var options []*parser.Option
	var rpcs []*parser.RPC
	for _, s := range src {
		switch t := s.(type) {
		case *parser.Option:
			options = append(options, t)
		case *parser.RPC:
			rpcs = append(rpcs, t)
		default:
			return nil, fmt.Errorf("invalid ServiceBody type %v of %v", t, s)
		}
	}
	return &ServiceBody{
		Options: options,
		RPCs:    rpcs,
	}, nil
}
