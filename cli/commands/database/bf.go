package database

import (
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/bloomfilter"
	"tprgdb/engine"
)

type BFCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
}

func (c *BFCommand) Execute() (commands.IContext, error) {
	return bloomfilter.NewBloomFilterContext(c.caller, c.dbEngine), nil
}

func NewBFCommand(caller commands.IContext, dbEngine *engine.KVEngine) *BFCommand {
	return &BFCommand{
		caller:   caller,
		dbEngine: dbEngine,
	}
}
