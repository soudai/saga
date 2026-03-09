package main

import (
	"io"
	"os"

	"github.com/soudai/saga/internal/cli"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cmd := cli.NewRootCommand(stdin, stdout, stderr)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		_, _ = io.WriteString(stderr, err.Error()+"\n")
		return 1
	}

	return 0
}
