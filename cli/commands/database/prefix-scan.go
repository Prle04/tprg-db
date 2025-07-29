package database

import (
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/scan"
	"tprgdb/engine"
)

type PrefixScanCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	prefix   string
	pageSize int
}

func (cmd *PrefixScanCommand) Execute() (commands.IContext, error) {
	cursor, err := cmd.dbEngine.PrefixScan(cmd.prefix, cmd.pageSize)
	if err != nil {
		return cmd.caller, err
	}
	scanContext := scan.NewScanContext(cmd.caller, cmd.dbEngine, cursor, "PS")
	return scanContext, nil
}

func NewPrefixScanCommand(caller commands.IContext, dbEngine *engine.KVEngine, prefix string, pageSize int) *PrefixScanCommand {
	return &PrefixScanCommand{
		caller:   caller,
		dbEngine: dbEngine,
		prefix:   prefix,
		pageSize: pageSize,
	}
}
