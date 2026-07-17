# template-renderer

Simple CLI wrapper for https://github.com/NikolaLohinski/gonja

```
Standalone golang wrapper for rendering Jinja2 templates

Usage:
  template-renderer [flags]

Flags:
  -d, --debug           Debug-level logging on stderr.
  -h, --help            help for template-renderer
  -i, --input string    Input template file (Required)
  -o, --output string   Outfile to write the outcome (Prints to stdout by default)
      --trace           Trace-level logs to /tmp/template-renderer.log (truncated each run).
      --version         Print the version and exit
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
