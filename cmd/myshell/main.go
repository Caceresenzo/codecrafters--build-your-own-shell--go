package main

import (
	"bufio"
	// Uncomment this block to pass the first stage
	// "fmt"
	"os"
)

func main() {
	f := bufio.NewWriter(os.Stdout)
	f.Write([]byte("$ "))
	f.Flush()

	// Wait for user input
	bufio.NewReader(os.Stdin).ReadString('\n')
}
