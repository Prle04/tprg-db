package commands

type BackCommand struct {
	caller IContext
}

func (cmd *BackCommand) Execute() (IContext, error) {
	// Simply return the caller context to go back
	err := cmd.caller.Close()
	if err != nil {
		return nil, err
	}
	return cmd.caller.GetOwner(), nil
}

func NewBackCommand(caller IContext) *BackCommand {
	return &BackCommand{
		caller: caller,
	}
}
