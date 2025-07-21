package countminsketch

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
	err := cmd.dbEngine.Probabilistic.CountMinSketch.Delete(cmd.name)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Printf("CMS '%s' deleted successfully.\n", cmd.name)
	return cmd.caller, nil
}

func NewDeleteCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string) *DeleteCommand {
	return &DeleteCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
	}
}
