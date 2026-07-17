package cmd

import (
	"os"
	"fmt"
	"strings"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/spf13/cobra"
)

var inputFile string
var outputFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "template-renderer",
	Short: "Standalone golang wrapper for rendering Jinja2 templates",

	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: renderTemplate,
}

func renderTemplate(cmd *cobra.Command, args []string) error {

	tpl, err := gonja.FromFile(inputFile)
	if err != nil {
		return err
	}

	data := map[string]any{}
	for _, element := range os.Environ() {
		variable := strings.Split(element, "=")
		data[variable[0]] = strings.Join(variable[1:], "=")
	}
	out, err := tpl.ExecuteToString(exec.NewContext(data))
	if err != nil {
		return err
	}
	if outputFile == "" {
		fmt.Printf(out)
	} else {
		err := os.WriteFile(outputFile, []byte(out), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input template file (Required)")
	rootCmd.MarkFlagRequired("input")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Outfile to write the outcome (Prints to stdout by default)")
}


