package meta_test

import (
	"testing"

	"github.com/thought-machine/go-protoparser/parser/meta"
)

func TestPosition_String(t *testing.T) {
	tests := []struct {
		name       string
		inputPos   meta.Position
		wantString string
	}{
		{
			name: "pos without Filename",
			inputPos: meta.Position{
				Offset: 0,
				Line:   1,
				Column: 1,
			},
			wantString: `<input>:1:1`,
		},
		{
			name: "pos with Filename",
			inputPos: meta.Position{
				Filename: "test.proto",
				Offset:   0,
				Line:     1,
				Column:   1,
			},
			wantString: `test.proto:1:1`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := test.inputPos.String()
			if got != test.wantString {
				t.Errorf("got %s, but want %s", got, test.wantString)
			}
		})
	}
}
