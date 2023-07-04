# template-renderer

Simple CLI wrapper for https://github.com/NikolaLohinski/gonja

```
Standalone golang wrapper for rendering Jinja2 templates

Usage:
  template-renderer [flags]
  template-renderer [command]

Available Commands:
  help        Help about any command
  version     Print version

Flags:
  -h, --help            help for template-renderer
  -i, --input string    Input template file (Required)
  -o, --output string   Outfile to write the outcome (Prints to stdout by default)

Use "template-renderer [command] --help" for more information about a command.
```

Usage examples:
- Templating to stdout:
```
./bin/template-renderer -i test.tpl
testing value: foo
```
- Templating to an output file:
```
./bin/template-renderer -i test.tpl -o test.txt
```
