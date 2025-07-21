package iter

import (
	"fmt"
	"tprgdb/cli/commands"
	engineScan "tprgdb/engine/api/scans"
)

type NextCommand struct {
	caller   commands.IContext
	iterator engineScan.IIterator
}

func (cmd *NextCommand) Execute() (commands.IContext, error) {

	// Fetch the next page of results
	item, err := cmd.iterator.Next()
	if err != nil {
		return cmd.caller, err
	}

	if item == nil {
		fmt.Printf("No more items to iterate.\n")
		return cmd.caller.GetOwner(), nil
	}

	// Print the results
	fmt.Printf("'%s' -> '%s' (%s)\n", item.Key, string(item.Value), item.GetTimestampString())

	return cmd.caller, nil
}

func NewNextCommand(caller commands.IContext, iterator engineScan.IIterator) *NextCommand {
	return &NextCommand{
		caller:   caller,
		iterator: iterator,
	}
}
