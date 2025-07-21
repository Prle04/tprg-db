package commands

type ErrorCommand struct {
	caller IContext
	err    error
}

func (c *ErrorCommand) Execute() (IContext, error) {
	// Return the current context with the error
	return c.caller, c.err
}

func NewErrorCommand(caller IContext, err error) *ErrorCommand {
	return &ErrorCommand{
		caller: caller,
		err:    err,
	}
}
