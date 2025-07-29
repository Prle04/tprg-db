package commands

type ICommand interface {
	Execute() (IContext, error)
}
