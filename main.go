package main

import (
	"os"

	"github.com/jholm117/hackerrank-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
