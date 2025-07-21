package database

import "tprgdb/cli/commands"

type LogoutCommand struct {
	caller commands.IContext
}

func (c *LogoutCommand) Execute() (commands.IContext, error) {
	// Close the current database context and return to the previous context
	if err := c.caller.Close(); err != nil {
		return c.caller, err
	}
	return c.caller.GetOwner(), nil // Returning the caller context's owner indicating we are logging out
}

func NewLogoutCommand(caller commands.IContext) *LogoutCommand {
	return &LogoutCommand{
		caller: caller,
	}
}
