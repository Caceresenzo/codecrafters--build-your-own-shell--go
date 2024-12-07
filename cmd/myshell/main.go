package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

func main() {
	builtins = make(map[string]BuiltinFunction)
	builtins["exit"] = builtin_exit
	builtins["echo"] = builtin_echo
	builtins["type"] = builtin_type
	builtins["pwd"] = builtin_pwd
	builtins["cd"] = builtin_cd

	for {
		line := read()

		if len(line) == 0 {
			break
		}

		eval(line)
	}
}
