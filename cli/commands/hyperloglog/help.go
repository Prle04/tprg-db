package hyperloglog

import (
	"fmt"
	"tprgdb/cli/commands"
)

type HelpCommand struct {
	caller commands.IContext
}

func (c *HelpCommand) Execute() (commands.IContext, error) {
	fmt.Println("HyperLogLog Commands:")
	fmt.Println("  new <name> <bits>       - Create a new HLL with the given name and number of bits")
	fmt.Println("  delete <name>           - Delete the specified HLL")
	fmt.Println("  insert <name> <value>   - Insert a value into the specified HLL")
	fmt.Println("  count <name>            - Estimate the cardinality of the specified HLL")
	fmt.Println("  clear                   - Clear the screen")
	fmt.Println("  exit                    - Exit HyperLogLog subsystem")
	fmt.Println("  help                    - Show this help message")

	return c.caller, nil
}

func NewHelpCommand(caller commands.IContext) *HelpCommand {
	return &HelpCommand{caller: caller}
}
