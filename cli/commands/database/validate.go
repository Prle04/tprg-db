package database

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type ValidateCommand struct {
	caller    commands.IContext
	dbEngine  *engine.KVEngine
	tableName string
}

func (cmd *ValidateCommand) Execute() (commands.IContext, error) {
	valid, err := cmd.dbEngine.Validate(cmd.tableName)
	if err != nil {
		return nil, err
	}
	if valid {
		fmt.Printf("table '%s' is valid.\n", cmd.tableName)
	} else {
		fmt.Printf("table '%s' is not valid.\n", cmd.tableName)
	}
	return cmd.caller, nil
}

func NewValidateCommand(caller commands.IContext, dbEngine *engine.KVEngine, tableName string) *ValidateCommand {
	return &ValidateCommand{
		caller:    caller,
		dbEngine:  dbEngine,
		tableName: tableName,
	}
}
