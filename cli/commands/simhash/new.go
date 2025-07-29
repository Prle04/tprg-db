package simhash

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type NewCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
	text     string
}

func (cmd *NewCommand) Execute() (commands.IContext, error) {
	err := cmd.dbEngine.Probabilistic.SimHash.Create(cmd.name, cmd.text)
	if err != nil {
		return cmd.caller, err
	}
	fmt.Printf("Fingerprint '%s' created  successfully.\n", cmd.name)
	return cmd.caller, nil
}

func NewNewCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string, text string) *NewCommand {
	return &NewCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
		text:     text,
	}
}
