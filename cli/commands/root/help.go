package root

import (
	"fmt"
	"tprgdb/cli/commands"
)

type HelpCommand struct {
	caller commands.IContext
}

func (c *HelpCommand) Execute() (commands.IContext, error) {
	fmt.Println("Root Commands:")
	fmt.Println("  new <db_name> <username> <password>   - Create a new database with given credentials")
	fmt.Println("  login <db_name> <username> <password> - Login to an existing database")
	fmt.Println("  quit                                  - Exit the application")
	fmt.Println("  clear                                 - Clear the screen")
	fmt.Println("  help                                  - Show this help message")

	return c.caller, nil
}

func NewHelpCommand(caller commands.IContext) *HelpCommand {
	return &HelpCommand{caller: caller}
}
