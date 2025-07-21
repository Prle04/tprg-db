package countminsketch

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type NewCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
	k        int
	m        int
}

func (cmd *NewCommand) Execute() (commands.IContext, error) {
	err := cmd.dbEngine.Probabilistic.CountMinSketch.Create(cmd.name, cmd.k, cmd.m)
	if err != nil {
		return nil, err
	}
	fmt.Printf("CMS '%s' created  successfully.\n", cmd.name)
	return cmd.caller, nil
}

func NewNewCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string, k, m int) *NewCommand {
	return &NewCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
		k:        k,
		m:        m,
	}
}
