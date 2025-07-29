package scan

import (
	"fmt"
	"tprgdb/cli/commands"
)

type HelpCommand struct {
	caller commands.IContext
}

func (c *HelpCommand) Execute() (commands.IContext, error) {
	fmt.Println("Scan Commands:")
	fmt.Println("  page <pageNumber>   - Jump to the specified page number")
	fmt.Println("  page-next           - Go to the next page")
	fmt.Println("  page-prev           - Go to the previous page")
	fmt.Println("  clear               - Clear the screen")
	fmt.Println("  exit                - Exit the scan context")
	fmt.Println("  help                - Show this help message")

	return c.caller, nil
}

func NewHelpCommand(caller commands.IContext) *HelpCommand {
	return &HelpCommand{caller: caller}
}
