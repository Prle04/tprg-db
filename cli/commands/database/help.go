package database

import (
	"fmt"
	"tprgdb/cli/commands"
)

type HelpCommand struct {
	caller commands.IContext
}

func (c *HelpCommand) Execute() (commands.IContext, error) {
	fmt.Println("Database Context Commands:")
	fmt.Println("  get <key>                       - Get the value for a key")
	fmt.Println("  get <key> -f <filepath>         - Get the value for a key and write to a file")
	fmt.Println("  get-many <k1> ... <kN>          - Get values for multiple keys")
	fmt.Println("  get-many <k1>... -f <f1>...     - Get values and write each to its file")
	fmt.Println("  set <key> <value>               - Set a value for a key")
	fmt.Println("  set <key> -f <filepath>         - Set value for key from file")
	fmt.Println("  delete <key>                    - Delete a key")
	fmt.Println("  validate <tableName>            - Validate table by name")
	fmt.Println("  prefix-scan <prefix> <pageSize> - Page through keys with a given prefix")
	fmt.Println("  range-scan <start> <end> <ps>   - Page through keys in a range")
	fmt.Println("  prefix-iter <prefix>            - Iterate keys with a given prefix")
	fmt.Println("  range-iter <start> <end>        - Iterate keys in a range")
	fmt.Println("  bf                              - Enter Bloom Filter subsystem")
	fmt.Println("  hll                             - Enter HyperLogLog subsystem")
	fmt.Println("  cms                             - Enter Count-Min Sketch subsystem")
	fmt.Println("  sh                              - Enter Stable HyperLogLog subsystem")
	fmt.Println("  clear                           - Clear the screen")
	fmt.Println("  logout                          - Exit to login")
	fmt.Println("  new-user <username> <password>  - Create a new user")
	fmt.Println("  help                            - Show this help message")

	return c.caller, nil
}

func NewHelpCommand(caller commands.IContext) *HelpCommand {
	return &HelpCommand{caller: caller}
}
