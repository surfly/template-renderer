package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/spf13/cobra"
)

// traceLogPath is where --trace writes its log, truncated on every run.
const traceLogPath = "/tmp/template-renderer.log"

var inputFile string
var outputFile string
var debug bool
var trace bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "template-renderer",
	Short: "Standalone golang wrapper for rendering Jinja2 templates",
	RunE:  renderTemplate,
}

func renderTemplate(cmd *cobra.Command, args []string) error {
	level, tracePath := "", ""
	switch {
	case trace:
		level, tracePath = "trace", traceLogPath
	case debug:
		level = "debug"
	}
	cleanup := initLogger(tracePath, level)
	defer cleanup()

	tpl, err := gonja.FromFile(inputFile)
	if err != nil {
		return fmt.Errorf("load template %q: %w", inputFile, err)
	}
	L.Debug("loaded template", slog.String("file", inputFile))

	data := map[string]any{}
	for _, element := range os.Environ() {
		variable := strings.Split(element, "=")
		data[variable[0]] = strings.Join(variable[1:], "=")
	}
	L.Debug("collected template context from environment", slog.Int("variables", len(data)))

	out, err := tpl.ExecuteToString(exec.NewContext(data))
	if err != nil {
		return fmt.Errorf("render template %q: %w", inputFile, err)
	}
	if outputFile == "" {
		fmt.Print(out)
		return nil
	}
	if err := os.WriteFile(outputFile, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write output %q: %w", outputFile, err)
	}
	L.Debug("wrote rendered output", slog.String("file", outputFile), slog.Int("bytes", len(out)))
	return nil
}

// Execute runs the root command and exits with a non-zero status on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.Flags().Bool("version", false, "Print the version and exit")
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input template file (Required)")
	cobra.CheckErr(rootCmd.MarkFlagRequired("input"))
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Outfile to write the outcome (Prints to stdout by default)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Debug-level logging on stderr.")
	rootCmd.PersistentFlags().BoolVar(&trace, "trace", false, "Trace-level logs to "+traceLogPath+" (truncated each run).")
}
