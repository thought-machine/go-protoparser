package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/thought-machine/go-protoparser/internal/lexer"
	"github.com/thought-machine/go-protoparser/internal/util_test"
	"github.com/thought-machine/go-protoparser/parser"
	"github.com/thought-machine/go-protoparser/parser/meta"
)

func TestParser_ParseField(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		permissive bool
		wantField  *parser.Field
		wantErr    bool
	}{
		{
			name:    "parsing an empty",
			wantErr: true,
		},
		{
			name:    "parsing an invalid; without fieldNumber",
			input:   "foo.bar nested_message = ;",
			wantErr: true,
		},
		{
			name:    "parsing an invalid; string fieldNumber",
			input:   "foo.bar nested_message = a;",
			wantErr: true,
		},
		{
			name:  "parsing an excerpt from the official reference",
			input: "foo.bar nested_message = 2;",
			wantField: &parser.Field{
				Type:        "foo.bar",
				FieldName:   "nested_message",
				FieldNumber: "2",
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
		{
			name:  "parsing another excerpt from the official reference",
			input: "repeated int32 samples = 4 [packed=true];",
			wantField: &parser.Field{
				IsRepeated:  true,
				Type:        "int32",
				FieldName:   "samples",
				FieldNumber: "4",
				FieldOptions: []*parser.FieldOption{
					{
						OptionName: "packed",
						Constant:   "true",
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
		{
			name:  "parsing fieldOptions",
			input: "repeated int32 samples = 4 [packed=true, required=false];",
			wantField: &parser.Field{
				IsRepeated:  true,
				Type:        "int32",
				FieldName:   "samples",
				FieldNumber: "4",
				FieldOptions: []*parser.FieldOption{
					{
						OptionName: "packed",
						Constant:   "true",
					},
					{
						OptionName: "required",
						Constant:   "false",
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
		{
			name:       "parsing deprecation syntax",
			input:      "string my_field = 7 [(release.field)={notice_version:{major:1,minor:7},release_version:{major:3},change_type:FIELD_REMOVAL,description:\"This field's functionality is to be replaced by my_new_field\"}];",
			permissive: true,
			wantField: &parser.Field{
				Type:        "string",
				FieldName:   "my_field",
				FieldNumber: "7",
				FieldOptions: []*parser.FieldOption{
					{
						OptionName: "(release.field)",
						Constant:   "{notice_version:{major:1,minor:7},release_version:{major:3},change_type:FIELD_REMOVAL,description:\"This field's functionality is to be replaced by my_new_field\"}",
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
		{
			name:       "parsing fieldOption constant with { by permissive mode. Required by go-proto-validators",
			input:      "int64 display_order = 1 [(validator.field) = {int_gt: 0}];",
			permissive: true,
			wantField: &parser.Field{
				Type:        "int64",
				FieldName:   "display_order",
				FieldNumber: "1",
				FieldOptions: []*parser.FieldOption{
					{
						OptionName: "(validator.field)",
						Constant:   "{int_gt:0}",
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
		{
			name:       "parsing fieldOption constant with { and , by permissive mode. Required by go-proto-validators",
			input:      `string email = 2 [(validator.field) = {length_gt: 0, length_lt: 1025},(validator.field) = {regex: "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}"}];`,
			permissive: true,
			wantField: &parser.Field{
				Type:        "string",
				FieldName:   "email",
				FieldNumber: "2",
				FieldOptions: []*parser.FieldOption{
					{
						OptionName: "(validator.field)",
						Constant:   "{length_gt:0,length_lt:1025}",
					},
					{
						OptionName: "(validator.field)",
						Constant:   `{regex:"[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}"}`,
					},
				},
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
			p := parser.NewParser(lexer.NewLexer(strings.NewReader(test.input)), parser.WithPermissive(test.permissive))
			got, err := p.ParseField()
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

			if !reflect.DeepEqual(got, test.wantField) {
				t.Errorf("got %v, but want %v", util_test.PrettyFormat(got), util_test.PrettyFormat(test.wantField))
			}

			if !p.IsEOF() {
				t.Errorf("got not eof, but want eof")
			}
		})
	}

}
