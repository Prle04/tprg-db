package database

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type SetCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	key      string
	value    []byte
}

func (cmd *SetCommand) Execute() (commands.IContext, error) {
	err := cmd.dbEngine.Set(cmd.key, cmd.value)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Key '%s' set successfully.\n", cmd.key)
	return cmd.caller, nil
}

func NewSetCommand(caller commands.IContext, dbEngine *engine.KVEngine, key string, value []byte) *SetCommand {
	return &SetCommand{
		caller:   caller,
		dbEngine: dbEngine,
		key:      key,
		value:    value,
	}
}
