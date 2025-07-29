package database

import (
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/simhash"
	"tprgdb/engine"
)

type SHCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
}

func (c *SHCommand) Execute() (commands.IContext, error) {
	return simhash.NewSimHashContext(c.caller, c.dbEngine), nil
}

func NewSHCommand(caller commands.IContext, dbEngine *engine.KVEngine) *SHCommand {
	return &SHCommand{
		caller:   caller,
		dbEngine: dbEngine,
	}
}
