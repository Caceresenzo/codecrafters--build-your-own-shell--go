package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type BuiltinFunction func([]string)

var builtins map[string]BuiltinFunction

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

func builtin_exit(_ []string) {
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

func builtin_pwd(_ []string) {
	current, _ := os.Getwd()
	fmt.Fprintf(os.Stdout, "%s\n", current)
}

func builtin_cd(arguments []string) {
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
			fmt.Fprintf(os.Stdout, "cd: $HOME is not set")
		} else {
			absolute = fmt.Sprintf("%s/%s", HOME, path[1:])
		}
	}

	if len(absolute) == 0 {
		return
	}

	if err := os.Chdir(absolute); errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stdout, "cd: %s: No such file or directory\n", path)
	}
}
