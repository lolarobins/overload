package input

import (
	"bufio"
	"os"
	"strings"

	"lolarobins.ca/overload/log"
)

type CommandFunc func([]string)

type Command struct {
	Command     string
	Description string
	Args        string
	Function    CommandFunc
}

var commands map[string]*Command = make(map[string]*Command)
var active bool = true

func Init() {
	Command{
		Function: func([]string) {
			log.Info("Showing available commands")

			for _, command := range commands {
				log.Info(command.Command + command.Args + " > " + command.Description)
			}
		},
		Command:     "help",
		Description: "List available commands",
	}.Register()

	Command{
		Function: func([]string) {
			active = false
		},
		Command:     "quit",
		Description: "Exit the program",
	}.Register()
}

func AcceptInput() {
	reader := bufio.NewReader(os.Stdin)

	for active {
		input, _ := reader.ReadString('\n')
		input = strings.Replace(input, "\n", "", -1)

		command := commands[strings.ToLower(strings.Split(input, " ")[0])]

		if command != nil {
			command.Function(strings.Split(input, " "))
		} else {
			log.Info("Command not found, type 'help' to see available commands")
		}
	}
}

func (c Command) Register() {
	commands[strings.ToLower(c.Command)] = &c
}
