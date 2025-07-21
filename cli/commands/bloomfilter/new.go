package bloomfilter

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type NewCommand struct {
	caller           commands.IContext
	dbEngine         *engine.KVEngine
	name             string
	numberOfElements int
	errorRate        float64
}

func (cmd *NewCommand) Execute() (commands.IContext, error) {
	err := cmd.dbEngine.Probabilistic.BloomFilter.Create(cmd.name, cmd.numberOfElements, cmd.errorRate)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Bloom filter '%s' created  successfully.\n", cmd.name)
	return cmd.caller, nil
}

func NewNewCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string, numberOfElements int, errorRate float64) *NewCommand {
	return &NewCommand{
		caller:           caller,
		dbEngine:         dbEngine,
		name:             name,
		numberOfElements: numberOfElements,
		errorRate:        errorRate,
	}
}
