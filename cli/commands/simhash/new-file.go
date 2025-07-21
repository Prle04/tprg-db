package simhash

import (
	"fmt"
	"os"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type NewFileCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	name     string
	filename string
}

func (cmd *NewFileCommand) Execute() (commands.IContext, error) {
	text, err := os.ReadFile(cmd.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", cmd.filename, err)
	}
	err = cmd.dbEngine.Probabilistic.SimHash.Create(cmd.name, string(text))
	if err != nil {
		return cmd.caller, err
	}
	fmt.Printf("Fingerprint '%s' created  successfully.\n", cmd.name)
	return cmd.caller, nil
}

func NewNewFileCommand(caller commands.IContext, dbEngine *engine.KVEngine, name string, filename string) *NewFileCommand {
	return &NewFileCommand{
		caller:   caller,
		dbEngine: dbEngine,
		name:     name,
		filename: filename,
	}
}
