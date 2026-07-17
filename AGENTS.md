# AGENTS.md

> See [AGENTS.universal.md](./AGENTS.universal.md) and [AGENTS.go.md](./AGENTS.go.md) for universal conventions.
> Refresh: `make standards`

---

## Overview

template-renderer is a standalone CLI that renders Jinja2 templates using the
[gonja](https://github.com/NikolaLohinski/gonja) library. All environment
variables are exposed as the template context, which makes it suitable for
generating configuration files in deployment scripts and containers.

---

## Architecture

```
main.go            Entry point; calls cmd.Execute
cmd/
  root.go          Root command: flags and template rendering
  logger.go        slog setup wired to the --debug/--trace flags
  version.go       Version variable stamped at build time via ldflags
  root_test.go     Table-driven tests for template rendering
```

---

## Build & Run

```bash
make check      # format → vet → build → test → lint
make build      # binaries for all platforms in bin/, symlink in repo root
```

Smoke test:

```bash
echo 'value: {{ FOO }}' > /tmp/smoke.tpl
FOO=bar ./template-renderer -i /tmp/smoke.tpl
# value: bar
```

---

## Design Decisions

- Environment variables are the only template context — there is no data file
  or CLI argument input.
- Environment values are split on the first "=" only, so values containing "="
  survive intact.
- CGO is disabled in builds so the binaries are static.

---

## Gotchas

- All template context values are strings; numeric operations in templates
  need explicit conversion filters.
