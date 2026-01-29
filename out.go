package main

import (
	"fmt"
	"os"
)

// prepareOut prepares the output directory for writing.
func prepareOut(out string, force bool) {
	st, err := os.Stat(out)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(out, 0o750); err != nil {
				fmt.Fprintln(os.Stderr, "mkdir out error:", err)
				os.Exit(1)
			}
			return
		}
		fmt.Fprintln(os.Stderr, "stat out error:", err)
		os.Exit(1)
	}

	if !st.IsDir() {
		fmt.Fprintln(os.Stderr, "out exists and is not a directory:", out)
		os.Exit(2)
	}

	ents, err := os.ReadDir(out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "readdir out error:", err)
		os.Exit(1)
	}

	if len(ents) == 0 {
		return
	}

	if !force {
		fmt.Fprintln(os.Stderr, "out directory is not empty (use --force):", out)
		os.Exit(2)
	}

	if err := os.RemoveAll(out); err != nil {
		fmt.Fprintln(os.Stderr, "failed to remove out:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(out, 0o750); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir out error:", err)
		os.Exit(1)
	}
}
