package main

import (
	"strings"

	"github.com/spf13/pflag"

	"fmt"
	"os"

	"github.com/paulhammond/cftest/internal/cftest"
)

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: cftest [function] [test...]\n")
		pflag.PrintDefaults()
	}

	pflag.Parse()
	args := pflag.Args()
	if len(args) < 2 {
		pflag.Usage()
		os.Exit(2)
	}

	runner, err := parseFunc(args[0])

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Testing %s:\n", runner.Name())

	tests, err := cftest.ReadTests(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	for _, t := range tests {
		r, err := cftest.RunTest(runner, t)

		if err != nil {
			fmt.Printf("ERROR: %s\n%s\n", t.Filename, err)
		}
		if r.OK {
			fmt.Printf("OK:    %s (%d)\n", t.Filename, r.Utilization)
		} else {
			fmt.Printf("FAIL:  %s (%d)\n%s\n", t.Filename, r.Utilization, r.Failure)
		}

	}
}

func parseFunc(name string) (cftest.Runner, error) {
	parts := strings.SplitN(name, ":", 2)

	if len(parts) > 1 && (parts[1] == "DEVELOPMENT" || parts[1] == "LIVE") {
		return cftest.NewCloudFrontRunner(parts[0], parts[1])
	}

	return nil, fmt.Errorf("function not found")
}
