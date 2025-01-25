package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

type AutocompleteResult int

const (
	AutocompleteNone AutocompleteResult = iota
	AutocompleteFound
	AutocompleteMore
)

func autocompletePrint(line *string, candidate string) {
	os.Stdout.WriteString(candidate)
	*line += candidate

	os.Stdout.WriteString(" ")
	*line += " "
}

func autocomplete(line *string) AutocompleteResult {
	var candidates []string

	for name := range builtins {
		if strings.HasPrefix(name, *line) {
			candidate := name[len(*line):]
			candidates = append(candidates, candidate)
		}
	}

	if len(candidates) == 0 {
		return AutocompleteNone
	}

	if len(candidates) == 1 {
		candidate := candidates[0]
		autocompletePrint(line, candidate)

		return AutocompleteNone
	}

	fmt.Printf("%d %v", len(candidates), candidates)
	panic("TODO")
	// return AutocompleteMore
}

type ReadResult int

const (
	ReadResultQuit ReadResult = iota
	ReadResultEmpty
	ReadResultContent
)

func read() (string, ReadResult) {
	os.Stdout.Write([]byte{'$', ' '})

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

	var line = ""

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
			result := autocomplete(&line)

			switch result {
			case AutocompleteNone:
				break
			case AutocompleteFound:
				break
			case AutocompleteMore:
				break
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
	parsedLine := parseArgv(line)

	io, valid := OpenIo(parsedLine.redirects)
	if !valid {
		return
	}

	defer io.Close()

	arguments := parsedLine.arguments
	program := arguments[0]

	if builtin, found := builtins[program]; found {
		builtin(arguments, io)
		return
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
