package scan

import (
	"fmt"
	"tprgdb/cli/commands"
	engineScan "tprgdb/engine/api/scans"
)

type PagePrevCommand struct {
	caller commands.IContext
	cursor engineScan.ICursor
}

func (cmd *PagePrevCommand) Execute() (commands.IContext, error) {

	// Fetch the previous page of results
	results, err := cmd.cursor.PreviousPage()
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

func NewPagePrevCommand(caller commands.IContext, cursor engineScan.ICursor) *PagePrevCommand {
	return &PagePrevCommand{
		caller: caller,
		cursor: cursor,
	}
}
