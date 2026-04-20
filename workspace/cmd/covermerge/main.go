package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type modeType string

const (
	modeSet    modeType = "set"
	modeCount  modeType = "count"
	modeAtomic modeType = "atomic"
)

func parseMode(line string) (modeType, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "mode: ") {
		return "", false
	}
	m := strings.TrimSpace(strings.TrimPrefix(line, "mode: "))
	switch modeType(m) {
	case modeSet, modeCount, modeAtomic:
		return modeType(m), true
	default:
		return modeType(m), true
	}
}

func mergeCount(existing int64, incoming int64, mode modeType) int64 {
	if mode == modeSet {
		if existing > 0 || incoming > 0 {
			return 1
		}
		return 0
	}
	return existing + incoming
}

func readProfile(path string, expectedMode *modeType, blocks map[string]int64) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if m, ok := parseMode(line); ok {
			if expectedMode != nil {
				if *expectedMode == "" {
					*expectedMode = m
				} else if *expectedMode != m {
					return fmt.Errorf("mode mismatch in %s: got %q, expected %q", path, m, *expectedMode)
				}
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 3 {
			return fmt.Errorf("invalid coverprofile line in %s:%d: %q", path, lineNo, line)
		}

		numStmts, err := strconv.Atoi(fields[1])
		if err != nil || numStmts < 0 {
			return fmt.Errorf("invalid numStmts in %s:%d: %q", path, lineNo, fields[1])
		}

		count, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid count in %s:%d: %q", path, lineNo, fields[2])
		}

		key := fields[0] + " " + fields[1]
		mode := modeCount
		if expectedMode != nil && *expectedMode != "" {
			mode = *expectedMode
		}

		blocks[key] = mergeCount(blocks[key], count, mode)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func writeMerged(w io.Writer, mode modeType, blocks map[string]int64) error {
	if mode == "" {
		return errors.New("missing coverage mode")
	}
	if _, err := fmt.Fprintf(w, "mode: %s\n", mode); err != nil {
		return err
	}

	keys := make([]string, 0, len(blocks))
	for k := range blocks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%s %d\n", k, blocks[k]); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var outPath string
	flag.StringVar(&outPath, "out", "", "output coverprofile path (defaults to stdout)")
	flag.Parse()

	profiles := flag.Args()
	if len(profiles) == 0 {
		fmt.Fprintln(os.Stderr, "usage: covermerge [-out path] <profile1> <profile2> ...")
		os.Exit(2)
	}

	blocks := make(map[string]int64, 1024)
	var mode modeType

	for _, p := range profiles {
		if err := readProfile(p, &mode, blocks); err != nil {
			fmt.Fprintf(os.Stderr, "covermerge: %v\n", err)
			os.Exit(1)
		}
	}

	var w io.Writer = os.Stdout
	var outFile *os.File
	if outPath != "" {
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "covermerge: %v\n", err)
			os.Exit(1)
		}
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "covermerge: %v\n", err)
			os.Exit(1)
		}
		outFile = f
		defer outFile.Close()
		w = outFile
	}

	if err := writeMerged(w, mode, blocks); err != nil {
		fmt.Fprintf(os.Stderr, "covermerge: %v\n", err)
		os.Exit(1)
	}
}
