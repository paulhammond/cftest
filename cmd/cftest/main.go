package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/pflag"

	"github.com/paulhammond/cftest/internal/cftest"
)

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: cftest [function] [test...]\n")
		pflag.PrintDefaults()
	}

	var help = pflag.BoolP("help", "h", false, "show help")
	pflag.CommandLine.MarkHidden("help")

	pflag.Parse()
	args := pflag.Args()

	if len(args) < 2 || *help {
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

	maxPathLen := 20
	for _, t := range tests {
		if len(t.Filename) > maxPathLen {
			maxPathLen = len(t.Filename)
		}
	}
	fmtPath := fmt.Sprintf("%%-%ds ", maxPathLen)

	for _, t := range tests {
		r, err := cftest.RunTest(runner, t)
		path := fmt.Sprintf(fmtPath, t.Filename)
		utilization := color.HiBlackString(fmt.Sprintf("(%d%%)", r.Utilization))

		if err != nil {
			fmt.Printf("%s %s %s\n%s\n\n", color.RedString("✘"), path, color.RedString("ERROR"), err)
			continue
		}
		if r.OK {
			fmt.Printf("%s %s %s    %s\n", color.GreenString("✔"), path, color.GreenString("ok"), utilization)
		} else {
			fmt.Printf("%s %s %s  %s\n%s\n\n", color.RedString("✘"), path, color.RedString("FAIL"), utilization, colorDiff(r.Failure))
		}

	}
}

func colorDiff(diff string) string {
	diff = strings.ReplaceAll(diff, "(-got +want)", fmt.Sprintf("(%s %s)", color.RedString("-got"), color.GreenString("+want")))
	lines := strings.Split(diff, "\n")
	for i, l := range lines {
		if l[0] == '+' {
			lines[i] = color.GreenString(l)
		}
		if l[0] == '-' {
			lines[i] = color.RedString(l)
		}
	}
	return strings.Join(lines, "\n")
}

func parseFunc(name string) (cftest.Runner, error) {
	parts := strings.SplitN(name, ":", 2)

	if len(parts) > 1 && (parts[1] == "DEVELOPMENT" || parts[1] == "LIVE") {
		return cftest.NewCloudFrontRunner(parts[0], parts[1])
	}

	return nil, fmt.Errorf("function not found")
}
