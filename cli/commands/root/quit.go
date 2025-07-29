package root

import (
	"tprgdb/cli/commands"
)

type QuitCommand struct {
	caller commands.IContext
}

func (c *QuitCommand) Execute() (commands.IContext, error) {
	// Close the current context and return nil to indicate quitting
	if err := c.caller.Close(); err != nil {
		return c.caller, err
	}
	return nil, nil // Returning nil indicates that we are quitting the CLI
}

func NewQuitCommand(caller commands.IContext) *QuitCommand {
	return &QuitCommand{
		caller: caller,
	}
}
