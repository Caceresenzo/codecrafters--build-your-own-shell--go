package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "$ ")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
