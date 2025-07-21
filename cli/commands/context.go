package commands

import "tprgdb/cli/parser"

type IContext interface {
	GetCommand(*parser.ParsedCommand) ICommand
	GetOwner() IContext
	GetPrompt() string
	Close() error
}
