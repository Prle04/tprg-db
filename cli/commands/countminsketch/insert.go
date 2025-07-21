package countminsketch

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type InsertCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
	value    []byte
}

func (cmd *InsertCommand) Execute() (commands.IContext, error) {
	err := cmd.dbEngine.Probabilistic.CountMinSketch.Insert(cmd.name, cmd.value)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Printf("Value '%s' inserted into CMS '%s' successfully.\n", cmd.value, cmd.name)
	return cmd.caller, nil
}

func NewInsertCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string, value []byte) *InsertCommand {
	return &InsertCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
		value:    value,
	}
}
