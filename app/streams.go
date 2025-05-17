package main

import (
	"fmt"
	"os"
)

type Io = interface {
	Output() *os.File
	Error() *os.File
	Close()
}

type BasicIo struct {
	output *os.File
	error  *os.File
}

func (io *BasicIo) Output() *os.File {
	if io.output != nil {
		return io.output
	} else {
		return os.Stdout
	}
}

func (io *BasicIo) Error() *os.File {
	if io.error != nil {
		return io.error
	} else {
		return os.Stderr
	}
}

func (io *BasicIo) Close() {
	if io.output != nil {
		io.output.Close()
		io.output = nil
	}

	if io.error != nil {
		io.error.Close()
		io.error = nil
	}
}

func OpenIo(redirects []redirect) (Io, bool) {
	var output *os.File = nil
	var error *os.File = nil

	for _, redirect := range redirects {
		flag := os.O_CREATE | os.O_WRONLY
		if redirect.append {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}

		file, err := os.OpenFile(redirect.path, flag, 0664)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shell: %s: %s\n", redirect.path, err.Error())
			return &BasicIo{nil, nil}, false
		}

		if redirect.streamName == standardOutput {
			if output != nil {
				output.Close()
			}

			output = file
		} else if redirect.streamName == standardError {
			if error != nil {
				error.Close()
			}

			error = file
		} else {
			file.Close()
		}
	}

	return &BasicIo{output, error}, true
}
