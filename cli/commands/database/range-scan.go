package database

import (
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/scan"
	"tprgdb/engine"
)

type RangeScanCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	start    string
	end      string
	pageSize int
}

func (cmd *RangeScanCommand) Execute() (commands.IContext, error) {
	cursor, err := cmd.dbEngine.RangeScan(cmd.start, cmd.end, cmd.pageSize)
	if err != nil {
		return nil, err
	}
	scanContext := scan.NewScanContext(cmd.caller, cmd.dbEngine, cursor, "RS")
	return scanContext, nil
}

func NewRangeScanCommand(caller commands.IContext, dbEngine *engine.KVEngine, start, end string, pageSize int) *RangeScanCommand {
	return &RangeScanCommand{
		caller:   caller,
		dbEngine: dbEngine,
		start:    start,
		end:      end,
		pageSize: pageSize,
	}
}
