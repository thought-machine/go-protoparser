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

func TestParser_ParseOption(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		permissive bool
		wantOption *parser.Option
		wantErr    bool
	}{
		{
			name:    "parsing an empty",
			wantErr: true,
		},
		{
			name:    "parsing an invalid; without option",
			input:   `java_package = "com.example.foo";`,
			wantErr: true,
		},
		{
			name:    "parsing an invalid; without =",
			input:   `option java_package "com.example.foo";`,
			wantErr: true,
		},
		{
			name:    "parsing an invalid; without ;",
			input:   `option java_package = "com.example.foo"`,
			wantErr: true,
		},
		{
			name:  "parsing an excerpt from the official reference",
			input: `option java_package = "com.example.foo";`,
			wantOption: &parser.Option{
				OptionName: "java_package",
				Constant:   `"com.example.foo"`,
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
					LastPos: meta.Position{
						Offset: 39,
						Line:   1,
						Column: 40,
					},
				},
			},
		},
		{
			name:  "parsing another excerpt from the official reference",
			input: `option (my_option).a = true;`,
			wantOption: &parser.Option{
				OptionName: "(my_option).a",
				Constant:   `true`,
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
					LastPos: meta.Position{
						Offset: 27,
						Line:   1,
						Column: 28,
					},
				},
			},
		},
		{
			name:  "parsing fullIdent",
			input: `option java_package.baz.bar = "com.example.foo";`,
			wantOption: &parser.Option{
				OptionName: "java_package.baz.bar",
				Constant:   `"com.example.foo"`,
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 0,
						Line:   1,
						Column: 1,
					},
					LastPos: meta.Position{
						Offset: 47,
						Line:   1,
						Column: 48,
					},
				},
			},
		},
		{
			name: `parsing "{" ident ":" constant { ident ":" constant } "}" by permissive mode.`,
			input: `
option (google.api.http) = {
    get: "/v1/projects/{project_id}/aggregated/addresses"
    rest_collection: "projects.addresses"
};`,
			permissive: true,
			wantOption: &parser.Option{
				OptionName: "(google.api.http)",
				Constant:   "",
				Endpoint: &parser.CloudEndpoint{
					Fields: []*parser.EndpointFieldOption{
						{
							OptionName: "get",
							Constant:   `"/v1/projects/{project_id}/aggregated/addresses"`,
							Meta: meta.Meta{
								Pos: meta.Position{
									Offset: 34,
									Line:   3,
									Column: 5,
								},
							},
						},
						{
							OptionName: "rest_collection",
							Constant:   `"projects.addresses"`,
							Meta: meta.Meta{
								Pos: meta.Position{
									Offset: 92,
									Line:   4,
									Column: 5,
								},
							},
						},
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 1,
						Line:   2,
						Column: 1,
					},
					LastPos: meta.Position{
						Offset: 131,
						Line:   5,
						Column: 2,
					},
				},
			},
		},
		{
			name: `parsing "{" ident ":" constant { "," ident ":" constant } "}" by permissive mode.`,
			input: `
option (google.api.http) = {
    post: "/v1/resources",
    body: "resource",
    rest_method_name: "insert"
};`,
			permissive: true,
			wantOption: &parser.Option{
				OptionName: "(google.api.http)",
				Constant:   "",
				Endpoint: &parser.CloudEndpoint{
					Fields: []*parser.EndpointFieldOption{
						{
							OptionName: "post",
							Constant:   "\"/v1/resources\"",
							Meta: meta.Meta{
								Pos: meta.Position{
									Offset: 34,
									Line:   3,
									Column: 5,
								},
							},
						},
						{
							OptionName: "body",
							Constant:   "\"resource\"",
							Meta: meta.Meta{
								Pos: meta.Position{
									Offset: 61,
									Line:   4,
									Column: 5,
								},
							},
						},
						{
							OptionName: "rest_method_name",
							Constant:   "\"insert\"",
							Meta: meta.Meta{
								Pos: meta.Position{
									Offset: 83,
									Line:   5,
									Column: 5,
								},
							},
						},
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 1,
						Line:   2,
						Column: 1,
					},
					LastPos: meta.Position{
						Offset: 111,
						Line:   6,
						Column: 2,
					},
				},
			},
		},
		{
			name: "parses multiline string literal in multi-option annotation",
			input: `
option (google.api.http) = {
    post: "/v1/resources",
    body: "res"
		      "ource",
    rest_method_name: "insert"
};`,
			permissive: true,
			wantOption: &parser.Option{
				OptionName: "(google.api.http)",
				Constant:   `{post:"/v1/resources",body:"resource",rest_method_name:"insert"}`,
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 1,
						Line:   2,
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
			got, err := p.ParseOption()
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

			if !reflect.DeepEqual(got, test.wantOption) {
				t.Errorf("got %v, but want %v", util_test.PrettyFormat(got), util_test.PrettyFormat(test.wantOption))
			}

			if !p.IsEOF() {
				t.Errorf("got not eof, but want eof")
			}
		})
	}

}
