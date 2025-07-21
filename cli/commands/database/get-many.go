package database

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
)

type GetManyCommand struct {
	caller   commands.IContext
	keys     []string
	dbEngine *engine.KVEngine
}

func (c *GetManyCommand) Execute() (commands.IContext, error) {
	items, err := c.dbEngine.GetMany(c.keys)
	if err != nil {
		return c.caller, err
	}
	for i, item := range items {
		if item == nil {
			fmt.Printf("'%s' doesn't exist\n", c.keys[i])
		} else {
			fmt.Printf("'%s' -> '%s' (%s)\n", item.Key, string(item.GetValue()), item.GetTimestampString())
		}
	}
	// Return the current context
	return c.caller, nil
}

func NewGetManyCommand(caller commands.IContext, dbEngine *engine.KVEngine, keys []string) *GetManyCommand {
	return &GetManyCommand{
		caller:   caller,
		dbEngine: dbEngine,
		keys:     keys,
	}
}
