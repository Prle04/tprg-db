package database

import (
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/hyperloglog"
	"tprgdb/engine"
)

type HLLCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
}

func (c *HLLCommand) Execute() (commands.IContext, error) {
	return hyperloglog.NewHyperLogLogContext(c.caller, c.dbEngine), nil
}

func NewHLLCommand(caller commands.IContext, dbEngine *engine.KVEngine) *HLLCommand {
	return &HLLCommand{
		caller:   caller,
		dbEngine: dbEngine,
	}
}
