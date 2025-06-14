package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type BuiltinFunction func([]string, Io)

var builtins map[string]BuiltinFunction

func locate(program string) (string, bool) {
	if strings.HasPrefix(program, "/") {
		if _, err := os.Stat(program); err == nil {
			return program, true
		}
	}

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

func builtin_exit(_ []string, _ Io) {
	os.Exit(0)
}

func builtin_echo(arguments []string, io Io) {
	parts := arguments[1:]
	line := strings.Join(parts, " ")
	fmt.Fprintf(io.Output(), "%s\n", line)
}

func builtin_type(arguments []string, io Io) {
	program := arguments[1]

	if _, found := builtins[program]; found {
		fmt.Fprintf(io.Output(), "%s is a shell builtin\n", program)
		return
	}

	if path, found := locate(program); found {
		fmt.Fprintf(io.Output(), "%s is %s\n", program, path)
		return
	}

	fmt.Fprintf(io.Output(), "%s: not found\n", program)
}

func builtin_pwd(_ []string, io Io) {
	current, _ := os.Getwd()
	fmt.Fprintf(io.Output(), "%s\n", current)
}

func builtin_cd(arguments []string, io Io) {
	absolute := ""
	path := arguments[1]

	if strings.HasPrefix(path, "/") {
		absolute = path
	} else if strings.HasPrefix(path, ".") {
		current, _ := os.Getwd()
		absolute = fmt.Sprintf("%s/%s", current, path)
	} else if strings.HasPrefix(path, "~") {
		HOME := os.Getenv("HOME")
		if len(HOME) == 0 {
			fmt.Fprintf(io.Error(), "cd: $HOME is not set")
		} else {
			absolute = fmt.Sprintf("%s/%s", HOME, path[1:])
		}
	}

	if len(absolute) == 0 {
		return
	}

	if err := os.Chdir(absolute); errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(io.Error(), "cd: %s: No such file or directory\n", path)
	}
}

func builtin_history(arguments []string, io Io) {
	var first string
	if len(arguments) > 1 {
		first = arguments[1]
	} else {
		first = ""
	}

	if first == "-r" {
		readHistoryFrom(arguments[2])
	} else if first == "-w" {
		writeHistoryTo(arguments[2])
	} else if first != "" {
		value, err := strconv.Atoi(arguments[1])

		if err != nil {
			fmt.Fprintf(io.Error(), "history: invalid number\n")
			return
		}

		start := len(history) - value
		printHistory(start, io)
	} else {
		printHistory(0, io)
	}
}
