package iter

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/cli/parser"
	"tprgdb/engine"
	engineScan "tprgdb/engine/api/scans"
)

// Assuming ICommand is defined in commands package
type IterContext struct {
	owner    commands.IContext
	iterator engineScan.IIterator
	dbEngine *engine.KVEngine
	iterType string
}

// Implement GetCommand
func (ctx *IterContext) GetCommand(cmd *parser.ParsedCommand) commands.ICommand {
	switch cmd.Command {

	case "next":
		return NewNextCommand(ctx, ctx.iterator)
	case "stop":
		return commands.NewBackCommand(ctx)
	case "clear":
		return commands.NewClearCommand(ctx)
	case "help":
		return NewHelpCommand(ctx)
	default:
		return nil
	}
}
func (ctx *IterContext) GetOwner() commands.IContext {
	return ctx.owner
}

// Implement GetPrompt
func (ctx *IterContext) GetPrompt() string {
	// Example prompt
	return fmt.Sprintf("%s@%s (%s)> ", ctx.dbEngine.GetUsername(), ctx.dbEngine.GetDBName(), ctx.iterType)
}

// Implement Close
func (ctx *IterContext) Close() error {
	// The owner is database context, so we close it
	ctx.dbEngine = nil
	return ctx.owner.Close() // Close the owner context
}

func NewIterContext(owner commands.IContext, dbEngine *engine.KVEngine, iterator engineScan.IIterator, iterType string) *IterContext {
	return &IterContext{
		owner:    owner,
		iterator: iterator,
		dbEngine: dbEngine,
		iterType: iterType,
	}
}
