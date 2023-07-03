package cmd

import (
	"os"
	"testing"
)

func TestUtils_RenderTemplate(t *testing.T) {
	var templateString string = "{{ TEST_STRING | replace('this', 'that', 1) }}"
	var expectedOutput string = "Something better than that"
	os.Setenv("TEST_STRING", "Something better than this")

	inputFile = "/tmp/template-renderer-test.in"
	outputFile = "/tmp/template-renderer-test.out"

	os.Remove(inputFile)
	err := os.WriteFile(inputFile, []byte(templateString), 0644)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	defer os.Remove(inputFile)

	renderTemplate(nil, []string{})
	defer os.Remove(outputFile)

	b, err := os.ReadFile(outputFile)

	if string(b) != expectedOutput {
		t.Errorf("Expected '%s', got '%s'", expectedOutput, string(b))
	}
}


