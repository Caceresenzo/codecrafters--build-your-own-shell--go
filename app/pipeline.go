package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runSingle(parsedLine parsedLine) CommandResult {
	io, valid := OpenIo(parsedLine.redirects)
	if !valid {
		return CommandResult{
			ExitCode:  1,
			ExitShell: false,
		}
	}

	defer io.Close()

	arguments := parsedLine.arguments
	program := arguments[0]

	if builtin, found := builtins[program]; found {
		return builtin(arguments, io)
	}

	if path, found := locate(program); found {
		command := exec.Cmd{
			Path:   path,
			Args:   arguments,
			Stdin:  os.Stdin,
			Stdout: io.Output(),
			Stderr: io.Error(),
		}

		command.Run()
		return CommandResult{
			ExitCode:  command.ProcessState.ExitCode(),
			ExitShell: false,
		}
	}

	fmt.Fprintf(os.Stdout, "%s: command not found\n", program)
	return CommandResult{
		ExitCode:  1,
		ExitShell: false,
	}
}

func runMultiple(parsedLines []parsedLine) {
	commands := make([]*exec.Cmd, 0)

	for index, parsedLine := range parsedLines {
		is_first := index == 0
		is_last := index == len(parsedLines)-1

		io, valid := OpenIo(parsedLine.redirects)
		if valid {
			defer io.Close()
		}

		arguments := parsedLine.arguments
		program := arguments[0]

		var command *exec.Cmd = nil

		if _, found := builtins[program]; found {
			command = &exec.Cmd{
				Path: shellProgramPath,
				Args: append([]string{shellProgramPath}, arguments...),
			}
		} else if path, found := locate(program); found {
			command = &exec.Cmd{
				Path: path,
				Args: arguments,
			}
		} else {
			command = &exec.Cmd{
				Path: shellProgramPath,
				Args: append([]string{shellProgramPath}, program),
			}
		}

		if !is_first {
			previousCommand := commands[index-1]

			in, _ := previousCommand.StdoutPipe()
			command.Stdin = in
		} else {
			command.Stdin = os.Stdin
		}

		if is_last {
			// TODO Redirect also other commands
			command.Stdout = io.Output()
		}

		command.Stderr = io.Error()

		commands = append(commands, command)
	}

	for index, command := range commands {
		if index != 0 {
			command.Start()
		}
	}

	commands[0].Run()

	for index, command := range commands {
		if index != 0 {
			command.Wait()
		}
	}
}
