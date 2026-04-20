package workspace

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkspaceModules(t *testing.T) {
	repoRoot, err := repoRootFromTestWD()
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	moduleDirs, err := parseGoWorkUseDirs(filepath.Join(repoRoot, "go.work"))
	if err != nil {
		t.Fatalf("failed to parse go.work: %v", err)
	}
	if len(moduleDirs) == 0 {
		t.Fatalf("no modules found in go.work")
	}

	// Avoid recursion: we run module tests from the root already.
	moduleDirs = filterOut(moduleDirs, "./")
	moduleDirs = filterOut(moduleDirs, ".")

	for _, moduleDir := range moduleDirs {
		moduleDir := moduleDir
		t.Run(moduleDir, func(t *testing.T) {
			// Keep module tests isolated to avoid overwhelming CI and local machines.
			t.Parallel()

			absDir := filepath.Join(repoRoot, moduleDir)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			cmd := exec.CommandContext(ctx, "go", "test", "./...")
			cmd.Dir = absDir

			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			err := cmd.Run()
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("timed out running go test in %s\n%s", moduleDir, out.String())
			}
			if err != nil {
				t.Fatalf("go test failed in %s: %v\n%s", moduleDir, err, out.String())
			}
		})
	}
}

func repoRootFromTestWD() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// This test lives in <repoRoot>/workspace.
	return filepath.Clean(filepath.Join(wd, "..")), nil
}

func parseGoWorkUseDirs(goWorkPath string) ([]string, error) {
	f, err := os.Open(goWorkPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inUseBlock := false
	var dirs []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "use") && strings.HasSuffix(line, "(") {
			inUseBlock = true
			continue
		}
		if inUseBlock {
			if line == ")" {
				inUseBlock = false
				continue
			}

			// Strip trailing comments.
			if idx := strings.Index(line, "//"); idx >= 0 {
				line = strings.TrimSpace(line[:idx])
			}
			line = strings.Trim(line, "\"")
			if line != "" {
				dirs = append(dirs, line)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Basic validation: ensure dirs look like relative paths.
	for _, d := range dirs {
		if filepath.IsAbs(d) {
			return nil, fmt.Errorf("go.work contains absolute path in use(): %q", d)
		}
	}
	return dirs, nil
}

func filterOut(items []string, value string) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		if it == value {
			continue
		}
		out = append(out, it)
	}
	return out
}
