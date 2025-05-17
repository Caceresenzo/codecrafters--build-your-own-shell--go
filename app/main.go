package main

import (
	"os"

	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
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
			os.Stdin.Read(buffer) // 'A' or 'B' or 'C' or 'D'

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

func eval(line string) {
	commands := parseArgv(line)

	if len(commands) == 1 {
		runSingle(commands[0])
	} else {
		runMultiple(commands)
	}
}

func main() {
	builtins = make(map[string]BuiltinFunction)
	builtins["exit"] = builtin_exit
	builtins["echo"] = builtin_echo
	builtins["type"] = builtin_type
	builtins["pwd"] = builtin_pwd
	builtins["cd"] = builtin_cd

	for {
		line, result := read()

		switch result {
		case ReadResultQuit:
			return
		case ReadResultEmpty:
			continue
		case ReadResultContent:
			eval(line)
		}
	}
}
