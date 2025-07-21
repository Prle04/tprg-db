package commands

type BackCommand struct {
	caller IContext
}

func (cmd *BackCommand) Execute() (IContext, error) {
	return cmd.caller.GetOwner(), nil
}

func NewBackCommand(caller IContext) *BackCommand {
	return &BackCommand{
		caller: caller,
	}
}
