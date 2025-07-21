package hyperloglog

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type CountCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
}

func (cmd *CountCommand) Execute() (commands.IContext, error) {
	count, err := cmd.dbEngine.Probabilistic.HyperLogLog.Count(cmd.name)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Println(count)
	return cmd.caller, nil
}

func NewCountCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string) *CountCommand {
	return &CountCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
	}
}
