package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func complete() {
	// when Bash calls the command to perform completion it will
	// set several environment variables including COMP_LINE.
	// If this variable is not set, then command is being invoked
	// normally and we can return.
	if _, ok := os.LookupEnv("COMP_LINE"); !ok {
		return
	}

	_ = filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		s, err := os.Stat(path)
		if err != nil {
			return nil
		}

		if !s.IsDir() {
			fmt.Println(path)
		}

		return nil
	})

	os.Exit(0)
}
