package countminsketch

import (
	"fmt"
	"strconv"
	"tprgdb/cli/commands"
	"tprgdb/cli/parser"
	"tprgdb/engine"
)

type CountMinSketchContext struct {
	owner    commands.IContext
	dbEngine *engine.KVEngine
}

func (ctx *CountMinSketchContext) GetCommand(cmd *parser.ParsedCommand) commands.ICommand {
	switch cmd.Command {
	case "new":
		if len(cmd.Args) != 3 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: new <name> <k> <m>"))
		}
		k, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			return commands.NewErrorCommand(ctx, fmt.Errorf("invalid k: %s", cmd.Args[1]))
		}
		m, err := strconv.Atoi(cmd.Args[2])
		if err != nil {
			return commands.NewErrorCommand(ctx, fmt.Errorf("invalid m: %s", cmd.Args[2]))
		}
		return NewNewCommand(ctx, ctx.dbEngine, cmd.Args[0], k, m)
	case "delete":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: delete <name>"))
		}
		return NewDeleteCommand(ctx, ctx.dbEngine, cmd.Args[0])
	case "insert":
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: insert <name> <value>"))
		}
		return NewInsertCommand(ctx, ctx.dbEngine, cmd.Args[0], []byte(cmd.Args[1]))
	case "count":
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: count <name> <value>"))
		}
		return NewCountCommand(ctx, ctx.dbEngine, cmd.Args[0], []byte(cmd.Args[1]))
	case "exit":
		return commands.NewBackCommand(ctx)
	case "clear":
		return commands.NewClearCommand(ctx)
	case "help":
		return NewHelpCommand(ctx)
	default:
		return nil
	}
}

func (ctx *CountMinSketchContext) GetOwner() commands.IContext {
	return ctx.owner
}

func (ctx *CountMinSketchContext) GetPrompt() string {
	return fmt.Sprintf("%s@%s (CMS)> ", ctx.dbEngine.GetUsername(), ctx.dbEngine.GetDBName())
}

func (ctx *CountMinSketchContext) Close() error {
	// The owner is database context, so we close it
	ctx.dbEngine = nil
	return ctx.owner.Close()
}

func NewCountMinSketchContext(owner commands.IContext, dbEngine *engine.KVEngine) *CountMinSketchContext {
	return &CountMinSketchContext{
		owner:    owner,
		dbEngine: dbEngine,
	}
}
