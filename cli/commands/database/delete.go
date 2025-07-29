package database

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type DeleteCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	key      string
}

func (cmd *DeleteCommand) Execute() (commands.IContext, error) {
	err := cmd.dbEngine.Delete(cmd.key)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Printf("Key '%s' deleted successfully.\n", cmd.key)
	return cmd.caller, nil
}

func NewDeleteCommand(caller commands.IContext, dbEngine *engine.KVEngine, key string) *DeleteCommand {
	return &DeleteCommand{
		caller:   caller,
		dbEngine: dbEngine,
		key:      key,
	}
}
