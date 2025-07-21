package database

import (
	"fmt"
	"os"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type GetToFileCommand struct {
	caller   commands.IContext
	FilePath string
	key      string
	dbEngine *engine.KVEngine
}

func (c *GetToFileCommand) Execute() (commands.IContext, error) {
	// Retrieve the value for the given key from the database
	value, err := c.dbEngine.Get(c.key)
	if err != nil {
		return c.caller, err
	}

	// Write the value to the specified file
	file, err := os.OpenFile(c.FilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return c.caller, fmt.Errorf("failed to open file '%s': %w", c.FilePath, err)
	}
	defer file.Close()
	_, err = file.Write(value.Value)
	if err != nil {
		return c.caller, fmt.Errorf("failed to write to file '%s': %w", c.FilePath, err)
	}

	fmt.Printf("Value for '%s' written to '%s'\n", c.key, c.FilePath)

	// Return the current context
	return c.caller, nil
}

func NewGetToFileCommand(caller commands.IContext, dbEngine *engine.KVEngine, key, filePath string) *GetToFileCommand {
	return &GetToFileCommand{
		caller:   caller,
		dbEngine: dbEngine,
		key:      key,
		FilePath: filePath,
	}
}
