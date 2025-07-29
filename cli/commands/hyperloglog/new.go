package hyperloglog

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type NewCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
	bits     int
}

func (cmd *NewCommand) Execute() (commands.IContext, error) {
	err := cmd.dbEngine.Probabilistic.HyperLogLog.Create(cmd.name, cmd.bits)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Printf("HLL '%s' created  successfully.\n", cmd.name)
	return cmd.caller, nil
}

func NewNewCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string, bits int) *NewCommand {
	return &NewCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
		bits:     bits,
	}
}
