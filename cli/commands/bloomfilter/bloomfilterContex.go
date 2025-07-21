package bloomfilter

import (
	"fmt"
	"strconv"
	"tprgdb/cli/commands"
	"tprgdb/cli/parser"
	"tprgdb/engine"
)

type BloomFilterContext struct {
	owner    commands.IContext
	dbEngine *engine.KVEngine
}

func (ctx *BloomFilterContext) GetCommand(cmd *parser.ParsedCommand) commands.ICommand {
	switch cmd.Command {
	case "new":
		if len(cmd.Args) != 3 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: new <name> <numberOfElements> <falsePositiveRate>"))
		}
		numberOfElements, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			return commands.NewErrorCommand(ctx, fmt.Errorf("invalid number of elements: %s", cmd.Args[1]))
		}
		errorRate, err := strconv.ParseFloat(cmd.Args[2], 64)
		if err != nil {
			return commands.NewErrorCommand(ctx, fmt.Errorf("invalid error rate: %s", cmd.Args[2]))
		}
		return NewNewCommand(ctx, ctx.dbEngine, cmd.Args[0], numberOfElements, errorRate)
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
	case "is-not-in":
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: is-not-in <name> <value>"))
		}
		return NewIsNotInCommand(ctx, ctx.dbEngine, cmd.Args[0], []byte(cmd.Args[1]))
	case "exit":
		return commands.NewBackCommand(ctx)
	default:
		return nil
	}
}

func (ctx *BloomFilterContext) GetOwner() commands.IContext {
	return ctx.owner
}

func (ctx *BloomFilterContext) GetPrompt() string {
	return fmt.Sprintf("%s@%s (BF)> ", ctx.dbEngine.GetUsername(), ctx.dbEngine.GetDBName())
}

func (ctx *BloomFilterContext) Close() error {
	// The owner is database context, so we close it
	ctx.dbEngine = nil
	return ctx.owner.Close()
}

func NewBloomFilterContext(owner commands.IContext, dbEngine *engine.KVEngine) commands.IContext {
	return &BloomFilterContext{
		owner:    owner,
		dbEngine: dbEngine,
	}
}
