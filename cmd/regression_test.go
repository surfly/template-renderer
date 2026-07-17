package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRegression renders synthetic templates covering every Jinja construct
// downstream deployment templates rely on, and compares the output
// byte-for-byte against goldens generated with template-renderer v1.4.0
// (gonja v1). See testdata/regression/README.md.
func TestRegression(t *testing.T) {
	root := filepath.Join("..", "testdata", "regression", "features")

	envFile, err := os.Open(filepath.Join(root, "env"))
	if err != nil {
		t.Fatalf("open env file: %s", err)
	}
	defer func() { _ = envFile.Close() }()
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			t.Fatalf("malformed env line: %q", line)
		}
		t.Setenv(key, value)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read env file: %s", err)
	}

	templates, err := filepath.Glob(filepath.Join(root, "*.j2"))
	if err != nil || len(templates) == 0 {
		t.Fatalf("no templates found in %s: %v", root, err)
	}
	for _, tpl := range templates {
		name := strings.TrimSuffix(filepath.Base(tpl), ".j2")
		t.Run(name, func(t *testing.T) {
			inputFile = tpl
			outputFile = filepath.Join(t.TempDir(), name)
			if err := renderTemplate(nil, nil); err != nil {
				t.Fatalf("render: %s", err)
			}
			got, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("read rendered output: %s", err)
			}
			want, err := os.ReadFile(filepath.Join(root, "golden", name))
			if err != nil {
				t.Fatalf("read golden output: %s", err)
			}
			if string(got) != string(want) {
				t.Errorf("rendered output differs from golden/%s:\n%s", name, firstDiff(got, want))
			}
		})
	}
}

func firstDiff(got, want []byte) string {
	offset, line := 0, 1
	for ; offset < len(got) && offset < len(want); offset++ {
		if got[offset] != want[offset] {
			break
		}
		if got[offset] == '\n' {
			line++
		}
	}
	window := func(b []byte) string {
		end := min(offset+80, len(b))
		return string(b[offset:end])
	}
	return fmt.Sprintf("line %d, byte %d\ngot:  %q\nwant: %q", line, offset, window(got), window(want))
}
