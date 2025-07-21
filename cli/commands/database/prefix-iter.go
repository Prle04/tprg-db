package database

import (
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/iter"
	"tprgdb/engine"
)

type PrefixIterCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	prefix   string
}

func (cmd *PrefixIterCommand) Execute() (commands.IContext, error) {
	iterator, err := cmd.dbEngine.PrefixIterate(cmd.prefix)
	if err != nil {
		return nil, err
	}
	scanContext := iter.NewIterContext(cmd.caller, cmd.dbEngine, iterator, "PI")
	return scanContext, nil
}

func NewPrefixIterCommand(caller commands.IContext, dbEngine *engine.KVEngine, prefix string) *PrefixIterCommand {
	return &PrefixIterCommand{
		caller:   caller,
		dbEngine: dbEngine,
		prefix:   prefix,
	}
}
