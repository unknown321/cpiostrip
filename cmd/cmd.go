package cmd

import (
	"cpiostrip/compare"
	"cpiostrip/strip"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func Run() {
	complete()
	var filename1 string
	var filename2 string
	var filename string
	var out string
	flag.StringVar(&filename1, "f1", "", "compare file 1")
	flag.StringVar(&filename2, "f2", "", "compare file 2")
	flag.StringVar(&filename, "in", "", "file to strip")
	flag.StringVar(&out, "out", "", "output file")

	sysOut := os.Stdout
	flag.CommandLine.SetOutput(sysOut)
	flag.Usage = func() {
		fmt.Fprintf(sysOut, "Usage: %s -in [FILE]\n", os.Args[0])
		fmt.Fprintf(sysOut, "   or: %s -in [FILE] -out [FILE]\n", os.Args[0])
		fmt.Fprintf(sysOut, "   or: %s -f1 [FILE] -f2 [FILE]\n", os.Args[0])
		fmt.Fprintf(sysOut, "\nReset modification time of cpio file and directory entries to Thu Jan  1 12:00:00 AM GMT 1970\n")
		fmt.Fprintf(sysOut, "Compare entries in archives and print fields that differ with -f1/-f2\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if filename2 != "" {
		if err := compare.Compare(filename1, filename2); err != nil {
			slog.Error("compare failed", "error", err.Error())
			os.Exit(1)
		}

		return
	}

	if len(os.Args) < 2 {
		slog.Error("no file provided")
		os.Exit(1)
	}

	if os.Args[1] == "" {
		slog.Error("no file provided")
		os.Exit(1)
	}

	if out == "" {
		out = filename
	}

	if err := strip.Strip(filename, out); err != nil {
		slog.Error("failed to strip", "error", err.Error())
		os.Exit(1)
	}
}
