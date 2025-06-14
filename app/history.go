package main

import (
	"bufio"
	"fmt"
	"os"
)

var history []string

func printHistory(start int, io Io) {
	for i, line := range history[start:] {
		fmt.Fprintf(io.Output(), "%5d  %s\n", start+i+1, line)
	}
}

func readHistoryFrom(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		history = append(history, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return false
	}

	return true
}
