package main

import (
	"os"
	"path/filepath"

	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

var (
	shellProgramPath string
)

func prompt() {
	os.Stdout.Write([]byte{'$', ' '})
}

type ReadResult int

const (
	ReadResultQuit ReadResult = iota
	ReadResultEmpty
	ReadResultContent
)

const (
	Up   = 'A'
	Down = 'B'
)

func changeLine(line *string, new string) {
	backspaces := ""
	spaces := ""

	for i := 0; i < len(*line); i++ {
		backspaces += "\b"
		spaces += " "
	}

	os.Stdout.WriteString(backspaces)
	os.Stdout.WriteString(spaces)
	os.Stdout.WriteString(backspaces)

	os.Stdout.WriteString(new)
	*line = new
}

func read() (string, ReadResult) {
	prompt()

	var stdinFd = os.Stdin.Fd()

	var previous unix.Termios
	if err := termios.Tcgetattr(stdinFd, &previous); err != nil {
		panic(err)
	}

	var new = unix.Termios(previous)
	new.Iflag &= unix.IGNCR  // ignore received CR
	new.Lflag ^= unix.ICANON // disable canonical mode
	new.Lflag ^= unix.ECHO   // disable echo of input
	// new.Lflag ^= unix.ISIG   // disable signal
	new.Cc[unix.VMIN] = 1
	new.Cc[unix.VTIME] = 0
	if err := termios.Tcsetattr(stdinFd, termios.TCSANOW, &new); err != nil {
		panic(err)
	}

	defer termios.Tcsetattr(stdinFd, termios.TCSANOW, &previous)

	historyLen := len(history)
	historyPosition := historyLen

	line := ""
	bell_rang := false

	buffer := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(buffer)
		if err != nil {
			return "", ReadResultQuit
		}

		character := buffer[0]

		switch character {
		case 0x4:
			return "", ReadResultQuit

		case '\r':
			fallthrough
		case '\n':
			os.Stdout.Write([]byte{'\r', '\n'})

			if len(line) == 0 {
				return "", ReadResultEmpty
			} else {
				return line, ReadResultContent
			}

		case '\t':
			result := autocomplete(&line, bell_rang)

			switch result {
			case AutocompleteNone:
				bell_rang = false
				bell()

			case AutocompleteFound:
				bell_rang = false

			case AutocompleteMore:
				bell_rang = true
				bell()
			}

		case 0x1b:
			os.Stdin.Read(buffer) // '['
			os.Stdin.Read(buffer)

			direction := buffer[0]

			if direction == Up && historyPosition != 0 {
				historyPosition--
				changeLine(&line, history[historyPosition])
			} else if direction == Down && historyPosition < historyLen {
				historyPosition++

				if historyPosition == historyLen {
					changeLine(&line, "")
				} else {
					changeLine(&line, history[historyPosition])
				}
			}

		case 0x7f:
			if len(line) != 0 {
				line = line[:len(line)-1]
				os.Stdout.Write([]byte{'\b', ' ', '\b'})
			}

		default:
			os.Stdout.Write(buffer)
			line += string(character)
		}
	}
}

type CommandResult struct {
	ExitCode  int
	ExitShell bool
}

func eval(line string) CommandResult {
	history = append(history, line)

	commands := parseArgv(line)

	if len(commands) == 1 {
		return runSingle(commands[0])
	} else {
		runMultiple(commands)
	}

	return CommandResult{
		ExitCode:  0,
		ExitShell: false,
	}
}

func shellMain() int {
	shellProgramPath, _ = filepath.Abs(os.Args[0])

	history = make([]string, 0)

	builtins = make(map[string]BuiltinFunction)
	builtins["exit"] = builtin_exit
	builtins["echo"] = builtin_echo
	builtins["type"] = builtin_type
	builtins["pwd"] = builtin_pwd
	builtins["cd"] = builtin_cd
	builtins["history"] = builtin_history

	initializeHistory()
	defer finalizeHistory()

	arguments := os.Args[1:]
	if len(arguments) != 0 {
		result := runSingle(parsedLine{
			arguments: arguments,
			redirects: make([]redirect, 0),
		})

		return result.ExitCode
	}

	for {
		line, result := read()

		switch result {
		case ReadResultQuit:
			break
		case ReadResultEmpty:
			continue
		case ReadResultContent:
			result := eval(line)

			if result.ExitShell {
				return result.ExitCode
			}
		}
	}
}

func main() {
	os.Exit(shellMain())
}
