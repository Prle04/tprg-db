package countminsketch

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type CountCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
	value    []byte
}

func (cmd *CountCommand) Execute() (commands.IContext, error) {
	count, err := cmd.dbEngine.Probabilistic.CountMinSketch.Count(cmd.name, cmd.value)
	if err != nil {
		return nil, err
	}
	fmt.Println(count)
	return cmd.caller, nil
}

func NewCountCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string, value []byte) *CountCommand {
	return &CountCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
		value:    value,
	}
}
