package main

import (
	"bufio"
	"fmt"
	"os"
)

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
	fmt.Fprintf(os.Stdout, "%s: command not found\n", line)
}

func main() {
	for {
		line := read()

		if len(line) == 0 {
			break
		}

		eval(line)
	}
}
