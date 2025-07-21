package hyperloglog

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type DeleteCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
}

func (cmd *DeleteCommand) Execute() (commands.IContext, error) {
	err := cmd.dbEngine.Probabilistic.HyperLogLog.Delete(cmd.name)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Printf("HLL '%s' deleted successfully.\n", cmd.name)
	return cmd.caller, nil
}

func NewDeleteCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string) *DeleteCommand {
	return &DeleteCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
	}
}
