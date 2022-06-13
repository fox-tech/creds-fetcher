package cli

type CLI struct {
	commands map[string]Command
}

type Command struct {
	name string
	doc  string
	f    CommandBody
}

type Flag struct {
	Name  string
	Doc   string
	Value interface{}
}

type FlagMap map[string]Flag

type CommandBody func(FlagMap) error
