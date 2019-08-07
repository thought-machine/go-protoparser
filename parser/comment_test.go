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

func TestComment_IsCStyle(t *testing.T) {
	tests := []struct {
		name         string
		inputComment *parser.Comment
		wantIsCStyle bool
	}{
		{
			name: "parsing a C-style comment",
			inputComment: &parser.Comment{
				Raw: `/*
comment
*/
`,
			},
			wantIsCStyle: true,
		},
		{
			name: "parsing a C++-style comment",
			inputComment: &parser.Comment{
				Raw: "// comment",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := test.inputComment.IsCStyle()
			if got != test.wantIsCStyle {
				t.Errorf("got %v, but want %v", got, test.wantIsCStyle)
			}
		})
	}
}

func TestComment_Lines(t *testing.T) {
	tests := []struct {
		name         string
		inputComment *parser.Comment
		wantLines    []string
	}{
		{
			name: "parsing a C-style comment",
			inputComment: &parser.Comment{
				Raw: `/*comment*/`,
			},
			wantLines: []string{
				"comment",
			},
		},
		{
			name: "parsing C-style comments",
			inputComment: &parser.Comment{
				Raw: `/* comment1
comment2
*/`,
			},
			wantLines: []string{
				" comment1",
				"comment2",
				"",
			},
		},
		{
			name: "parsing a C++-style comment",
			inputComment: &parser.Comment{
				Raw: "// comment",
			},
			wantLines: []string{
				" comment",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := test.inputComment.Lines()
			if !reflect.DeepEqual(got, test.wantLines) {
				t.Errorf("got %v, but want %v", util_test.PrettyFormat(got), util_test.PrettyFormat(test.wantLines))
			}
		})
	}
}

func TestParser_ParseComments(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantComments []*parser.Comment
	}{
		{
			name: "parsing an empty",
		},
		{
			name: "parsing a C++-style comment",
			input: `// comment
`,
			wantComments: []*parser.Comment{
				{
					Raw: `// comment`,
					Meta: meta.Meta{
						Pos: meta.Position{
							Offset: 0,
							Line:   1,
							Column: 1,
						},
					},
				},
			},
		},
		{
			name: "parsing C++-style comments",
			input: `// comment
// comment2
`,
			wantComments: []*parser.Comment{
				{
					Raw: `// comment`,
					Meta: meta.Meta{
						Pos: meta.Position{
							Offset: 0,
							Line:   1,
							Column: 1,
						},
					},
				},
				{
					Raw: `// comment2`,
					Meta: meta.Meta{
						Pos: meta.Position{
							Offset: 11,
							Line:   2,
							Column: 1,
						},
					},
				},
			},
		},
		{
			name: "parsing a C-style comment",
			input: `/*
comment
*/`,
			wantComments: []*parser.Comment{
				{
					Raw: `/*
comment
*/`,
					Meta: meta.Meta{
						Pos: meta.Position{
							Offset: 0,
							Line:   1,
							Column: 1,
						},
					},
				},
			},
		},
		{
			name: "parsing C-style comments",
			input: `/*
comment
*/
/*
comment2
*/`,
			wantComments: []*parser.Comment{
				{
					Raw: `/*
comment
*/`,
					Meta: meta.Meta{
						Pos: meta.Position{
							Offset: 0,
							Line:   1,
							Column: 1,
						},
					},
				},
				{
					Raw: `/*
comment2
*/`,
					Meta: meta.Meta{
						Pos: meta.Position{
							Offset: 14,
							Line:   4,
							Column: 1,
						},
					},
				},
			},
		},
		{
			name: "parsing a C-style comment and a C++-style comment",
			input: `/*
comment
*/

// comment2
`,
			wantComments: []*parser.Comment{
				{
					Raw: `/*
comment
*/`,
					Meta: meta.Meta{
						Pos: meta.Position{
							Offset: 0,
							Line:   1,
							Column: 1,
						},
					},
				},
				{
					Raw: `// comment2`,
					Meta: meta.Meta{
						Pos: meta.Position{
							Offset: 15,
							Line:   5,
							Column: 1,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			p := parser.NewParser(lexer.NewLexer(strings.NewReader(test.input)))
			got := p.ParseComments()

			if !reflect.DeepEqual(got, test.wantComments) {
				t.Errorf("got %v, but want %v", util_test.PrettyFormat(got), util_test.PrettyFormat(test.wantComments))
			}

			if !p.IsEOF() {
				t.Errorf("got not eof, but want eof")
			}
		})
	}
}
