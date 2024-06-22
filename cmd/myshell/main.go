package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type BuiltinFunction func([]string)

var builtins map[string]BuiltinFunction

func read() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, "$ ")

		line, err := reader.ReadString('\n')
		if err != nil {
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

	if builtin, found := builtins[program]; found {
		builtin(arguments)
		return
	}

	if path, found := locate(program); found {
		command := exec.Cmd{
			Path:   path,
			Args:   arguments,
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		command.Run()
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

func builtin_type(arguments []string) {
	program := arguments[1]

	if _, found := builtins[program]; found {
		fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", program)
		return
	}

	if path, found := locate(program); found {
		fmt.Fprintf(os.Stdout, "%s is %s\n", program, path)
		return
	}

	fmt.Fprintf(os.Stdout, "%s: not found\n", program)
}

func locate(program string) (string, bool) {
	PATH := os.Getenv("PATH")
	directories := strings.Split(PATH, ":")

	for _, directory := range directories {
		path := fmt.Sprintf("%s/%s", directory, program)

		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}

	return "", false
}

func main() {
	builtins = make(map[string]BuiltinFunction)
	builtins["exit"] = builtin_exit
	builtins["echo"] = builtin_echo
	builtins["type"] = builtin_type

	for {
		line := read()

		if len(line) == 0 {
			break
		}

		eval(line)
	}
}
