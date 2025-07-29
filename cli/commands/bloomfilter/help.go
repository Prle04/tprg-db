package bloomfilter

import (
	"fmt"
	"tprgdb/cli/commands"
)

type HelpCommand struct {
	caller commands.IContext
}

func (c *HelpCommand) Execute() (commands.IContext, error) {
	fmt.Println("BloomFilter Commands:")
	fmt.Println("  new <name> <numberOfElements> <falsePositiveRate>   - Create a new Bloom filter")
	fmt.Println("  delete <name>                                       - Delete an existing Bloom filter")
	fmt.Println("  insert <name> <value>                               - Insert a value into a Bloom filter")
	fmt.Println("  is-not-in <name> <value>                            - Check if a value is definitely not in the Bloom filter")
	fmt.Println("  clear                                               - Clear the screen")
	fmt.Println("  exit                                                - Go back to the previous context")
	fmt.Println("  help                       						   - Show this help message")
	return c.caller, nil
}

func NewHelpCommand(caller commands.IContext) *HelpCommand {
	return &HelpCommand{caller: caller}
}
