package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/soudai/saga/internal/version"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 1
	}

	switch strings.TrimSpace(args[0]) {
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, version.String())
		return 0
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		printUsage(stderr)
		return 1
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: saga <command>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  help      show this help message")
	fmt.Fprintln(w, "  version   print build information")
}
