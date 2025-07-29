package simhash

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type GetCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
}

func (cmd *GetCommand) Execute() (commands.IContext, error) {
	fg, err := cmd.dbEngine.Probabilistic.SimHash.Get(cmd.name)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Println(fg)
	return cmd.caller, nil
}

func NewGetCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string) *GetCommand {
	return &GetCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
	}
}
