package simhash

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type SimilarityCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name1    string
	name2    string
}

func (cmd *SimilarityCommand) Execute() (commands.IContext, error) {
	distance, err := cmd.dbEngine.Probabilistic.SimHash.Similarity(cmd.name1, cmd.name2)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Println(distance)
	return cmd.caller, nil
}

func NewSimilarityCommand(caller commands.IContext, dbEngine *engine.KVEngine, name1, name2 string) *SimilarityCommand {
	return &SimilarityCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name1:    name1,
		name2:    name2,
	}
}
