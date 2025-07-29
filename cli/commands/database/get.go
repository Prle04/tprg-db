package database

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type GetCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	key      string
}

func (c *GetCommand) Execute() (commands.IContext, error) {
	// Retrieve the value for the given key from the database
	value, err := c.dbEngine.Get(c.key)
	if err != nil {
		return c.caller, err
	}

	if value == nil {
		fmt.Printf("'%s' does not exist.\n", c.key)
		return c.caller, nil
	}

	fmt.Printf("'%s' -> '%s' (%s)\n", c.key, string(value.Value), value.GetTimestampString())

	// Return the current context
	return c.caller, nil
}

func NewGetCommand(caller commands.IContext, dbEngine *engine.KVEngine, key string) *GetCommand {
	return &GetCommand{
		caller:   caller,
		dbEngine: dbEngine,
		key:      key,
	}
}
