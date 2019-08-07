package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/thought-machine/go-protoparser/internal/lexer"
	"github.com/thought-machine/go-protoparser/parser"
	"github.com/thought-machine/go-protoparser/parser/meta"
)

func TestParser_ParseSyntax(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSyntax *parser.Syntax
		wantErr    bool
	}{
		{
			name:    "parsing an empty",
			wantErr: true,
		},
		{
			name:  "parsing an excerpt from the official reference",
			input: `syntax = "proto3";`,
			wantSyntax: &parser.Syntax{
				ProtobufVersion: "proto3",
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			p := parser.NewParser(lexer.NewLexer(strings.NewReader(test.input)))
			got, err := p.ParseSyntax()
			switch {
			case test.wantErr:
				if err == nil {
					t.Errorf("got err nil, but want err")
				}
				return
			case !test.wantErr && err != nil:
				t.Errorf("got err %v, but want nil", err)
				return
			}

			if !reflect.DeepEqual(got, test.wantSyntax) {
				t.Errorf("got %v, but want %v", got, test.wantSyntax)
			}

			if !p.IsEOF() {
				t.Errorf("got not eof, but want eof")
			}
		})
	}

}
