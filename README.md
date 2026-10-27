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

## Template context

Every environment variable is exposed as a template variable:

```
echo 'value: {{ FOO }}' > test.tpl
FOO=bar template-renderer -i test.tpl
value: bar
```

All context values are strings. The template's trailing newline is kept
(unlike stock Jinja, which strips it), so rendered config files end with
`\n` exactly when the template does.

`or` and `and` return one of their operands, as in Jinja, so
`{{ PORT or "8080" }}` renders `8080` when `PORT` is empty or unset.
Version 1.6.2 and earlier rendered `True` or `False` instead.

## Available filters

All [gonja v2 builtin filters](https://github.com/NikolaLohinski/gonja/blob/master/docs/filters.md)
(strict Jinja2 parity: `default`, `format`, `join`, `lower`, `replace`,
`trim`, `tojson`, ...) plus the following gonja v1 filters that were dropped
by the v2 migration and are restored by this wrapper:

| filter | usage | result |
| ------ | ----- | ------ |
| `fromjson` | `{% set cfg = CFG_JSON \| fromjson %}` | parses a JSON string into a value; numbers keep their literal form (`8080`, not `8080.0`). Errors on empty input, invalid JSON, or trailing data |
| `split` | `{{ ADDR \| split(':') }}` | splits a string on a delimiter into a list: `{{ (ADDR \| split(':'))[0] }}`, `{{ NAMES \| split(' ') \| join('\|') }}` |
| `unset` | `{{ cfg \| unset("key") }}` | returns a copy of a dict without the given key; missing keys are ignored |

Combined example, as used by deployment templates:

```jinja
{%- set BACKENDS = BACKENDS_JSON | default("{}") | fromjson -%}
{%- set BACKENDS = BACKENDS | unset("") -%}
{%- for id in BACKENDS -%}
    {%- set host = (BACKENDS[id] | split(':'))[0] -%}
node {{ id }} {{ host }}
{% endfor %}
```

## Tests

`make test` runs the unit tests plus a golden-file regression suite
(`testdata/regression/`) that pins rendering behavior across gonja upgrades.
This repository is public — keep all test data synthetic.
