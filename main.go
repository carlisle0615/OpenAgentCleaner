package main

import (
	"os"

	"github.com/carlisle0615/OpenAgentCleaner/internal/cleaner"
)

func main() {
	os.Exit(cleaner.Run(os.Args[1:], os.Stdout, os.Stderr))
}
