package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type BuiltinFunction func([]string)

var builtins map[string]BuiltinFunction

func read() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, "$ ")

		line, error := reader.ReadString('\n')
		if error != nil {
			return ""
		}

		line = line[:len(line)-1]

		if len(line) != 0 {
			return line
		}
	}
}

func eval(line string) {
	arguments := strings.Split(line, " ")
	program := arguments[0]

	builtin, found := builtins[program]
	if found {
		builtin(arguments)
		return
	}

	fmt.Fprintf(os.Stdout, "%s: command not found\n", program)
}

func builtin_exit(arguments []string) {
	os.Exit(0)
}

func builtin_echo(arguments []string) {
	parts := arguments[1:]
	line := strings.Join(parts, " ")
	fmt.Fprintf(os.Stdout, "%s\n", line)
}

func main() {
	builtins = make(map[string]BuiltinFunction)
	builtins["exit"] = builtin_exit
	builtins["echo"] = builtin_echo

	for {
		line := read()

		if len(line) == 0 {
			break
		}

		eval(line)
	}
}
