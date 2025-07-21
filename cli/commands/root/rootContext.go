package root

import (
	"errors"
	"tprgdb/cli/commands"
	"tprgdb/cli/parser"
)

type RootContext struct {
}

func NewRootContext() *RootContext {
	return &RootContext{}
}

func (rc *RootContext) GetCommand(cmd *parser.ParsedCommand) commands.ICommand {
	switch cmd.Command {
	case "quit":
		return NewQuitCommand(rc)
	case "new":
		if len(cmd.Args) != 3 {
			return commands.NewErrorCommand(rc, errors.New("usage: new-db <db_name> <username> <password>"))
		}
		return NewNewDBCommand(rc, cmd.Args[0], cmd.Args[1], cmd.Args[2])
	case "login":
		if len(cmd.Args) != 3 {
			return commands.NewErrorCommand(rc, errors.New("usage: login <db_name> <username> <password>"))
		}
		return NewLoginCommand(rc, cmd.Args[0], cmd.Args[1], cmd.Args[2])
	case "clear":
		return commands.NewClearCommand(rc)
	case "help":
		return NewHelpCommand(rc)
	default:
		return nil // No command found for the root context
	}
}

func (rc *RootContext) GetPrompt() string {
	return "~ "
}

func (rc *RootContext) Close() error {
	// No resources to close in the root context
	return nil
}

func (rc *RootContext) GetOwner() commands.IContext {
	return nil // Root context has no owner
}
