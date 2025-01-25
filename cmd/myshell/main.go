package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

func prompt() {
	os.Stdout.Write([]byte{'$', ' '})
}

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

func autocomplete(line *string, bell_rang bool) AutocompleteResult {
	var candidates []string

	for name := range builtins {
		if strings.HasPrefix(name, *line) {
			candidate := name[len(*line):]
			candidates = append(candidates, candidate)
		}
	}

	PATH := os.Getenv("PATH")
	for _, directory := range strings.Split(PATH, ":") {
		entries, err := os.ReadDir(directory)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, *line) {
				continue
			}

			path := strings.Join([]string{directory, name}, "/")

			stat, err := os.Stat(path)
			if err != nil || !stat.Mode().IsRegular() || stat.Mode().Perm()&0111 == 0 {
				continue
			}

			candidate := name[len(*line):]
			if !slices.Contains(candidates, candidate) {
				candidates = append(candidates, candidate)
			}
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

	if bell_rang {
		slices.SortFunc(candidates, func(left string, right string) int {
			left_length := len(left)
			right_length := len(right)

			if left_length != right_length {
				return left_length - right_length
			}

			return strings.Compare(left, right)
		})

		os.Stdout.WriteString("\n")

		for index, candidate := range candidates {
			if index != 0 {
				os.Stdout.WriteString("  ")
			}

			os.Stdout.WriteString(*line)
			os.Stdout.WriteString(candidate)
		}

		os.Stdout.WriteString("\n")
		prompt()
		os.Stdout.WriteString(*line)
	}

	return AutocompleteMore
}

func bell() {
	os.Stdout.Write([]byte{'\a'})
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
