package database

import (
	"fmt"
	"os"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type SetFileCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	key      string
	fileName string
}

func (cmd *SetFileCommand) Execute() (commands.IContext, error) {
	value, err := os.ReadFile(cmd.fileName)
	if err != nil {
		return cmd.caller, err
	}
	err = cmd.dbEngine.Set(cmd.key, value)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Printf("Key '%s' set successfully.\n", cmd.key)
	return cmd.caller, nil
}

func NewSetFileCommand(caller commands.IContext, dbEngine *engine.KVEngine, key string, filename string) *SetFileCommand {
	return &SetFileCommand{
		caller:   caller,
		dbEngine: dbEngine,
		key:      key,
		fileName: filename,
	}
}
