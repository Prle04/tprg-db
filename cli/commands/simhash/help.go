package simhash

import (
	"fmt"
	"tprgdb/cli/commands"
)

type HelpCommand struct {
	caller commands.IContext
}

func (c *HelpCommand) Execute() (commands.IContext, error) {
	fmt.Println("SimHash Commands:")
	fmt.Println("  new <name> <text>           - Create a new SimHash from text")
	fmt.Println("  new <name> -f <filename>    - Create a new SimHash from a file")
	fmt.Println("  get <name>                  - Retrieve the SimHash value by name")
	fmt.Println("  similarity <name1> <name2>  - Compute similarity between two SimHashes")
	fmt.Println("  delete <name>               - Delete a SimHash by name")
	fmt.Println("  clear                       - Clear the screen")
	fmt.Println("  exit                        - Exit the SimHash context")
	fmt.Println("  help                        - Show this help message")

	return c.caller, nil
}

func NewHelpCommand(caller commands.IContext) *HelpCommand {
	return &HelpCommand{caller: caller}
}
