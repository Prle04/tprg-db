package iter

import (
	"fmt"
	"tprgdb/cli/commands"
)

type HelpCommand struct {
	caller commands.IContext
}

func (c *HelpCommand) Execute() (commands.IContext, error) {
	fmt.Println("Iterator Commands:")
	fmt.Println("  next     - Advance to the next element in the iterator")
	fmt.Println("  stop     - Stop iteration and return to the previous context")
	fmt.Println("  clear    - Clear the screen")
	fmt.Println("  help     - Show this help message")

	return c.caller, nil
}

func NewHelpCommand(caller commands.IContext) *HelpCommand {
	return &HelpCommand{caller: caller}
}
