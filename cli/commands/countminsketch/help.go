package countminsketch

import (
	"fmt"
	"tprgdb/cli/commands"
)

type HelpCommand struct {
	caller commands.IContext
}

func (c *HelpCommand) Execute() (commands.IContext, error) {
	fmt.Println("CountMinSketch Commands:")
	fmt.Println("  new <name> <k> <m>         - Create a new Count-Min Sketch with k hash functions and m width")
	fmt.Println("  delete <name>              - Delete an existing Count-Min Sketch")
	fmt.Println("  insert <name> <value>      - Insert a value into the Count-Min Sketch")
	fmt.Println("  count <name> <value>       - Estimate the count of a value")
	fmt.Println("  clear                      - Clear the screen")
	fmt.Println("  exit                       - Go back to the previous context")
	fmt.Println("  help                       - Show this help message")
	return c.caller, nil
}

func NewHelpCommand(caller commands.IContext) *HelpCommand {
	return &HelpCommand{caller: caller}
}
