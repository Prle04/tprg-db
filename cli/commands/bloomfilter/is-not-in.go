package bloomfilter

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type IsNotInCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
	value    []byte
}

func (cmd *IsNotInCommand) Execute() (commands.IContext, error) {
	isNotIn, err := cmd.dbEngine.Probabilistic.BloomFilter.IsNotIn(cmd.name, cmd.value)
	if err != nil {
		return nil, err
	}
	if isNotIn {
		fmt.Printf("Value '%s' is not in bloom filter '%s'.\n", cmd.value, cmd.name)
	} else {
		fmt.Printf("Value '%s' is in bloom filter '%s'.\n", cmd.value, cmd.name)
	}
	return cmd.caller, nil
}

func NewIsNotInCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string, value []byte) *IsNotInCommand {
	return &IsNotInCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
		value:    value,
	}
}
