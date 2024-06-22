package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Fprint(os.Stdout, "$ ")

	line, error := reader.ReadString('\n')
	if error != nil {
		os.Exit(0)
	}

	line = line[:len(line)-1]
	fmt.Fprintf(os.Stdout, "%s: command not found\n", line)
}
