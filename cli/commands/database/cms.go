package database

import (
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/countminsketch"
	"tprgdb/engine"
)

type CMSCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
}

func (c *CMSCommand) Execute() (commands.IContext, error) {
	return countminsketch.NewCountMinSketchContext(c.caller, c.dbEngine), nil
}

func NewCMSCommand(caller commands.IContext, dbEngine *engine.KVEngine) *CMSCommand {
	return &CMSCommand{
		caller:   caller,
		dbEngine: dbEngine,
	}
}
