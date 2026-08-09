package main

import (
	"os"

	"github.com/RichardFlp/portmaster/internal/cli"
	"github.com/RichardFlp/portmaster/internal/version"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr, os.Stdin, version.Version))
}
