package main

import (
	"bufio"
	"fmt"
	"os"
)

var history []string
var lastAppendIndex int = 0

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

func writeHistoryTo(path string) bool {
	file, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return false
	}
	defer file.Close()

	for _, line := range history {
		file.WriteString(line)
		file.WriteString("\n")
	}

	return true
}

func appendHistoryTo(path string) bool {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return false
	}
	defer file.Close()

	for _, line := range history[lastAppendIndex:] {
		file.WriteString(line)
		file.WriteString("\n")
	}

	lastAppendIndex = len(history)

	return true
}
