package scan

import (
	"fmt"
	"tprgdb/cli/commands"
	engineScan "tprgdb/engine/api/scans"
)

type PageNextCommand struct {
	caller commands.IContext
	cursor engineScan.ICursor
}

func (cmd *PageNextCommand) Execute() (commands.IContext, error) {

	// Fetch the next page of results
	results, err := cmd.cursor.NextPage()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		fmt.Printf("No results found for page %d.\n", cmd.cursor.GetCurrentPage())
		return cmd.caller, nil
	}

	// Print the results
	fmt.Printf("Page %d:\n", cmd.cursor.GetCurrentPage())
	for _, result := range results {
		fmt.Printf("'%s' -> '%s' (%s)\n", result.Key, string(result.Value), result.GetTimestampString())
	}

	return cmd.caller, nil
}

func NewPageNextCommand(caller commands.IContext, cursor engineScan.ICursor) *PageNextCommand {
	return &PageNextCommand{
		caller: caller,
		cursor: cursor,
	}
}
