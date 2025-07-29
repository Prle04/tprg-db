package database

import (
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/iter"
	"tprgdb/engine"
)

type RangeIterCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	start    string
	end      string
}

func (cmd *RangeIterCommand) Execute() (commands.IContext, error) {
	iterator, err := cmd.dbEngine.RangeIterate(cmd.start, cmd.end)
	if err != nil {
		return cmd.caller, err
	}
	scanContext := iter.NewIterContext(cmd.caller, cmd.dbEngine, iterator, "RI")
	return scanContext, nil
}

func NewRangeIterCommand(caller commands.IContext, dbEngine *engine.KVEngine, start, end string) *RangeIterCommand {
	return &RangeIterCommand{
		caller:   caller,
		dbEngine: dbEngine,
		start:    start,
		end:      end,
	}
}
