package database

import (
	"fmt"
	"os"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type GetManyFileCommand struct {
	caller    commands.IContext
	dbEngine  *engine.KVEngine
	keys      []string
	filenames []string
}

func (cmd *GetManyFileCommand) Execute() (commands.IContext, error) {

	items, err := cmd.dbEngine.GetMany(cmd.keys)
	if err != nil {
		return nil, err
	}
	totalItems := 0
	writtenItems := 0
	var resultErr error = nil
	for i, item := range items {
		file, err := os.OpenFile(cmd.filenames[i], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			if resultErr == nil {
				resultErr = err
			}
			continue
		}
		if item == nil {
			file.Close()
			continue
		}
		totalItems++

		_, err = file.Write(item.Value)
		if err != nil {
			if resultErr == nil {
				resultErr = err
			}
			file.Close()
			continue
		}
		file.Close()
		writtenItems++
	}
	fmt.Printf("Total items found: %d, successfuly written: %d\n", totalItems, writtenItems)
	return cmd.caller, resultErr
}

func NewGetManyFileCommand(caller commands.IContext, dbEngine *engine.KVEngine, keys []string, filenames []string) *GetManyFileCommand {
	return &GetManyFileCommand{
		caller:    caller,
		dbEngine:  dbEngine,
		keys:      keys,
		filenames: filenames,
	}
}
