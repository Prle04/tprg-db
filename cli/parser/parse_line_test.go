package parser_test

import (
	"reflect"
	"testing"
	"tprgdb/cli/parser"
)

func TestParseInputLine(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantCommand  string
		wantArgs     []string
		wantFlags    map[string][]string
		expectingErr bool
	}{
		{
			name:        "simple command with args",
			input:       `set key "value with spaces"`,
			wantCommand: "set",
			wantArgs:    []string{"key", "value with spaces"},
			wantFlags:   map[string][]string{},
		},
		{
			name:        "flags with multiple values",
			input:       `get-many key1 key2 key3 -f file1 file2 file3`,
			wantCommand: "get-many",
			wantArgs:    []string{"key1", "key2", "key3"},
			wantFlags:   map[string][]string{"-f": {"file1", "file2", "file3"}},
		},
		{
			name:        "flag with quoted value",
			input:       `fingerprint name -f "file with spaces.txt"`,
			wantCommand: "fingerprint",
			wantArgs:    []string{"name"},
			wantFlags:   map[string][]string{"-f": {"file with spaces.txt"}},
		},
		{
			name:        "command with escaped quotes",
			input:       `insert filterName "value with \"escaped quotes\""`,
			wantCommand: "insert",
			wantArgs:    []string{"filterName", `value with "escaped quotes"`},
			wantFlags:   map[string][]string{},
		},
		{
			name:        "multiple flags and args",
			input:       `cmd arg1 arg2 -flag1 val1 val2 val3 -flag2 "another val" `,
			wantCommand: "cmd",
			wantArgs:    []string{"arg1", "arg2"},
			wantFlags: map[string][]string{
				"-flag1": {"val1", "val2", "val3"},
				"-flag2": {"another val"},
			},
		},
		{
			name:         "empty input",
			input:        ``,
			expectingErr: true,
		},
		{
			name:         "unclosed quote",
			input:        `set key "unclosed value`,
			expectingErr: true,
		},
		{
			name:        "flag without value",
			input:       `cmd -flag1`,
			wantCommand: "cmd",
			wantArgs:    []string{},
			wantFlags: map[string][]string{
				"-flag1": {},
			},
		},
		{
			name:        "positional args only",
			input:       `delete key1 key2 key3`,
			wantCommand: "delete",
			wantArgs:    []string{"key1", "key2", "key3"},
			wantFlags:   map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseInputLine(tt.input)
			if (err != nil) != tt.expectingErr {
				t.Fatalf("parseInputLine() error = %v, expectingErr %v", err, tt.expectingErr)
			}
			if err != nil {
				// error expected, no further checks
				return
			}

			if got.Command != tt.wantCommand {
				t.Errorf("Command = %v, want %v", got.Command, tt.wantCommand)
			}
			if len(got.Args) == 0 && len(tt.wantArgs) == 0 {
				return // both empty, nothing to check
			}
			if len(got.Args) != len(tt.wantArgs) {
				t.Errorf("Args length = %d, want %d", len(got.Args), len(tt.wantArgs))
			} else {
				for i := range got.Args {
					if got.Args[i] != tt.wantArgs[i] {
						t.Errorf("Arg[%d] = %v, want %v", i, got.Args[i], tt.wantArgs[i])
					}
				}
			}

			if len(got.Flags) != len(tt.wantFlags) {
				t.Errorf("Flags length = %d, want %d", len(got.Flags), len(tt.wantFlags))
			}
			if len(got.Flags) == 0 && len(tt.wantFlags) == 0 {
				return // both empty, nothing to check
			}
			for k, v := range tt.wantFlags {
				gotV, ok := got.Flags[k]
				if !ok || !reflect.DeepEqual(gotV, v) {
					t.Errorf("Flag %v = %v, want %v", k, gotV, v)
				}
			}
			for k := range got.Flags {
				if _, ok := tt.wantFlags[k]; !ok {
					t.Errorf("Unexpected flag %v in result", k)
				}
			}
		})
	}
}
