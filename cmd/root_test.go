package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
