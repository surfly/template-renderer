package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nikolalohinski/gonja/v2/builtins"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		env      map[string]string
		missing  bool
		want     string
		wantErr  string
	}{
		{
			name:     "environment variable substitution",
			template: "value: {{ TEST_VALUE }}",
			env:      map[string]string{"TEST_VALUE": "foo"},
			want:     "value: foo",
		},
		{
			name:     "trailing newline is kept",
			template: "value: {{ TEST_VALUE }}\n",
			env:      map[string]string{"TEST_VALUE": "foo"},
			want:     "value: foo\n",
		},
		{
			name:     "filter applied to variable",
			template: "{{ TEST_STRING | replace('this', 'that', 1) }}",
			env:      map[string]string{"TEST_STRING": "Something better than this"},
			want:     "Something better than that",
		},
		{
			name:     "value containing equals signs",
			template: "{{ TEST_EQ }}",
			env:      map[string]string{"TEST_EQ": "a=b=c"},
			want:     "a=b=c",
		},
		{
			name:     "fromjson object access",
			template: "{%- set D = TEST_JSON | fromjson -%}{{ D['host'] }}:{{ D['port'] }}",
			env:      map[string]string{"TEST_JSON": `{"host": "10.0.0.1", "port": 8080}`},
			want:     "10.0.0.1:8080",
		},
		{
			name:     "fromjson of undefined variable with default",
			template: `{%- set D = MISSING_VALUE | default("{}") | fromjson -%}{% if 'x' in D %}yes{% else %}no{% endif %}`,
			want:     "no",
		},
		{
			name:     "fromjson of invalid JSON",
			template: "{{ TEST_JSON | fromjson }}",
			env:      map[string]string{"TEST_JSON": "not json"},
			wantErr:  "render template",
		},
		{
			name:     "fromjson of JSON with trailing data",
			template: "{{ TEST_JSON | fromjson }}",
			env:      map[string]string{"TEST_JSON": "{} junk"},
			wantErr:  "render template",
		},
		{
			name:     "fromjson of empty string",
			template: "{{ TEST_JSON | fromjson }}",
			env:      map[string]string{"TEST_JSON": ""},
			wantErr:  "render template",
		},
		{
			name:     "fromjson of undefined variable",
			template: "{{ MISSING_VALUE | fromjson }}",
			wantErr:  "render template",
		},
		{
			name:     "fromjson with unexpected argument",
			template: "{{ TEST_JSON | fromjson('x') }}",
			env:      map[string]string{"TEST_JSON": "{}"},
			wantErr:  "render template",
		},
		{
			name:     "fromjson of errored expression",
			template: "{{ MISSING_VALUE.x | fromjson }}",
			wantErr:  "render template",
		},
		{
			name:     "split of errored expression",
			template: "{{ MISSING_VALUE.x | split(':') }}",
			wantErr:  "render template",
		},
		{
			name:     "unset of errored expression",
			template: "{{ MISSING_VALUE.x | unset('a') }}",
			wantErr:  "render template",
		},
		{
			name:     "split and join",
			template: "{{ TEST_SERVERS | split(' ') | join('|') }}",
			env:      map[string]string{"TEST_SERVERS": "a.example b.example"},
			want:     "a.example|b.example",
		},
		{
			name:     "split and index",
			template: "{{ (TEST_ADDR | split(':'))[0] }}",
			env:      map[string]string{"TEST_ADDR": "10.0.0.1:443"},
			want:     "10.0.0.1",
		},
		{
			name:     "split of undefined variable",
			template: "{{ MISSING_VALUE | split(':') }}",
			wantErr:  "render template",
		},
		{
			name:     "split without delimiter",
			template: "{{ TEST_ADDR | split }}",
			env:      map[string]string{"TEST_ADDR": "10.0.0.1:443"},
			wantErr:  "render template",
		},
		{
			name:     "unset of missing key keeps dict intact",
			template: `{%- set D = TEST_JSON | fromjson | unset("zz") -%}{{ D['a'] }}`,
			env:      map[string]string{"TEST_JSON": `{"a": "keep"}`},
			want:     "keep",
		},
		{
			name:     "unset of non-dict",
			template: "{{ TEST_ADDR | unset('a') }}",
			env:      map[string]string{"TEST_ADDR": "10.0.0.1:443"},
			wantErr:  "render template",
		},
		{
			name:     "unset without key",
			template: "{{ TEST_JSON | fromjson | unset }}",
			env:      map[string]string{"TEST_JSON": `{"a": "keep"}`},
			wantErr:  "render template",
		},
		{
			// Mirrors the multiserver-config preamble of downstream haproxy templates
			name: "multiserver config idiom",
			template: `{%- set CONFIG = MULTISERVER | default("{}") | fromjson -%}
{%- if '' in CONFIG -%}{%- set CONFIG['00'] = CONFIG[''] -%}{%- endif -%}
{%- set CONFIG = CONFIG | unset("") -%}
{%- set ns = namespace(value="") -%}
{%- for item in CONFIG -%}{%- set parts = CONFIG[item] | split(':') -%}{%- set ns.value = ns.value ~ " " ~ parts[0] -%}{%- endfor -%}
{{ ns.value | trim }}`,
			env:  map[string]string{"MULTISERVER": `{"": "10.0.0.1:443"}`},
			want: "10.0.0.1",
		},
		{
			name:    "missing input file",
			missing: true,
			wantErr: "load template",
		},
		{
			name:     "invalid template syntax",
			template: "{{ TEST_VALUE",
			wantErr:  "load template",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			inputFile = filepath.Join(dir, "input.tpl")
			outputFile = filepath.Join(dir, "output.txt")
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			if !tt.missing {
				if err := os.WriteFile(inputFile, []byte(tt.template), 0o644); err != nil {
					t.Fatalf("write template: %s", err)
				}
			}
			err := renderTemplate(nil, nil)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			b, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("read output: %s", err)
			}
			if string(b) != tt.want {
				t.Errorf("expected %q, got %q", tt.want, string(b))
			}
		})
	}
}

func writeTemplate(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	inputFile = filepath.Join(dir, "input.tpl")
	outputFile = ""
	if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write template: %s", err)
	}
}

func TestRenderTemplateToStdout(t *testing.T) {
	writeTemplate(t, "value: {{ TEST_VALUE }}")
	t.Setenv("TEST_VALUE", "foo")

	saved := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %s", err)
	}
	os.Stdout = w
	renderErr := renderTemplate(nil, nil)
	_ = w.Close()
	os.Stdout = saved

	if renderErr != nil {
		t.Fatalf("unexpected error: %s", renderErr)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %s", err)
	}
	if string(b) != "value: foo" {
		t.Errorf("expected %q, got %q", "value: foo", string(b))
	}
}

func TestRenderTemplateWriteError(t *testing.T) {
	writeTemplate(t, "static")
	outputFile = filepath.Join(t.TempDir(), "no-such-dir", "output.txt")

	err := renderTemplate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "write output") {
		t.Errorf("expected write output error, got %v", err)
	}
}

func TestRenderTemplateDebugLogging(t *testing.T) {
	writeTemplate(t, "static")
	outputFile = filepath.Join(filepath.Dir(inputFile), "output.txt")
	debug = true
	t.Cleanup(func() { debug = false })

	if err := renderTemplate(nil, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestRenderTemplateTraceLogging(t *testing.T) {
	writeTemplate(t, "static")
	outputFile = filepath.Join(filepath.Dir(inputFile), "output.txt")
	trace = true
	t.Cleanup(func() {
		trace = false
		_ = os.Remove(traceLogPath)
	})

	if err := renderTemplate(nil, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if _, err := os.Stat(traceLogPath); err != nil {
		t.Errorf("expected trace log at %s: %s", traceLogPath, err)
	}
}

func TestRegisterFiltersDuplicate(t *testing.T) {
	err := registerFilters(builtins.Filters)
	if err == nil || !strings.Contains(err.Error(), "register filter") {
		t.Errorf("expected duplicate registration error, got %v", err)
	}
}

func TestExecute(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.tpl")
	output := filepath.Join(dir, "output.txt")
	if err := os.WriteFile(input, []byte("value: {{ TEST_VALUE }}"), 0o644); err != nil {
		t.Fatalf("write template: %s", err)
	}
	t.Setenv("TEST_VALUE", "foo")
	rootCmd.SetArgs([]string{"-i", input, "-o", output})

	Execute()

	b, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %s", err)
	}
	if string(b) != "value: foo" {
		t.Errorf("expected %q, got %q", "value: foo", string(b))
	}
}
