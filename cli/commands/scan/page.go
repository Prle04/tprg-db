package scan

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine/api/scans"
)

type PageCommand struct {
	caller     commands.IContext
	cursor     scans.ICursor
	pageNumber int
}

func (cmd *PageCommand) Execute() (commands.IContext, error) {

	// Fetch the next page of results
	results, err := cmd.cursor.GetPage(cmd.pageNumber)
	if err != nil {
		return cmd.caller, err
	}

	if len(results) == 0 {
		fmt.Printf("No results found for page %d.\n", cmd.pageNumber)
		return cmd.caller, nil
	}

	// Print the results 
	fmt.Printf("Page %d:\n", cmd.pageNumber)
	for _, result := range results {
		fmt.Printf("'%s' -> '%s' (%s)\n", result.Key, string(result.Value), result.GetTimestampString())
	}

	return cmd.caller, nil
}

func NewPageCommand(caller commands.IContext, cursor scans.ICursor, pageNumber int) *PageCommand {
	return &PageCommand{
		caller:     caller,
		cursor:     cursor,
		pageNumber: pageNumber,
	}
}
