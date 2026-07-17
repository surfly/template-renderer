# Rendering regression suite

`TestRegression` renders each `features/*.j2` with the context from
`features/env` and compares the result byte-for-byte against
`features/golden/<name>`. This pins rendering behavior — filters, whitespace
control, trailing newline, dict iteration order, literal `%` handling — so
template-renderer (and its gonja dependency) can be upgraded without silently
changing the output of downstream deployment templates.

The templates are synthetic but exercise every Jinja construct those
deployment templates rely on: `fromjson` / `split` / `unset` (gonja v1
filters restored in `cmd/filters.go`), `default`, `format`, `lower`, `trim`,
`replace`, `join`, `namespace()` accumulation, subscript assignment, dict
membership and nested access, macros, `loop.index`, and `~` concatenation.
All values in `env` are synthetic — never place production configuration,
secrets, or internal hostnames here: this repository is public.

## Refreshing goldens

Only regenerate after verifying the new output is an intended change:

```bash
cd testdata/regression/features
while IFS= read -r line; do [ -n "$line" ] && export "$line"; done < env
for t in *.j2; do template-renderer -i "$t" -o "golden/$(basename "$t" .j2)"; done
```

The current goldens were generated with template-renderer v1.4.0 (gonja v1),
the last release before the gonja v2 migration, to guarantee upgrade parity.
