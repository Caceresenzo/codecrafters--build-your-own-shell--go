package main

import (
	"fmt"
	"io"
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

	var line string
	for {
		_, err := fmt.Fscanf(file, "%s\n", &line)
		if err != nil {
			if err == io.EOF {
				break
			}

			return false
		}
		history = append(history, line)
	}

	return true
}
