package parser

import (
	"errors"
	"strings"
	"unicode"
)

type ParsedCommand struct {
	Command string
	Args    []string
	Flags   map[string][]string
}

// ParseInputLine parses input into command, args, and flags with multiple values support.
func ParseInputLine(input string) (*ParsedCommand, error) {
	var result ParsedCommand
	result.Flags = make(map[string][]string)

	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, errors.New("empty input")
	}

	result.Command = tokens[0]
	i := 1

	for i < len(tokens) {
		t := tokens[i]
		if strings.HasPrefix(t, "-") {
			// flag found
			flagName := t
			i++
			// gather all subsequent tokens that are NOT flags as values for this flag
			var values []string
			for i < len(tokens) && !strings.HasPrefix(tokens[i], "-") {
				values = append(values, tokens[i])
				i++
			}
			// store even if no values (empty slice means flag without args)
			result.Flags[flagName] = values
		} else {
			// positional argument
			result.Args = append(result.Args, t)
			i++
		}
	}

	return &result, nil
}

// tokenize is the same as before: handles quoted strings and escapes
func tokenize(input string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inQuote := false
	var quoteChar rune

	escapeNext := false

	for _, r := range input {
		switch {
		case escapeNext:
			current.WriteRune(r)
			escapeNext = false

		case r == '\\':
			escapeNext = true

		case inQuote:
			if r == quoteChar {
				inQuote = false
				tokens = append(tokens, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}

		case unicode.IsSpace(r):
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}

		case r == '"' || r == '\'':
			inQuote = true
			quoteChar = r

		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		if inQuote {
			return nil, errors.New("unclosed quote in input")
		}
		tokens = append(tokens, current.String())
	}

	if inQuote {
		return nil, errors.New("unclosed quote in input")
	}

	return tokens, nil
}
