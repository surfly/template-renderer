package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nikolalohinski/gonja/v2/builtins"
	"github.com/nikolalohinski/gonja/v2/exec"
)

// Filters that existed in gonja v1 but were dropped from gonja v2 builtins.
// Deployment templates (haproxy, varnish, ...) still rely on them.
func init() {
	if err := registerFilters(builtins.Filters); err != nil {
		panic(err)
	}
}

func registerFilters(filters *exec.FilterSet) error {
	for name, filter := range map[string]exec.FilterFunction{
		"fromjson": filterFromJSON,
		"split":    filterSplit,
		"unset":    filterUnset,
	} {
		if err := filters.Register(name, filter); err != nil {
			return fmt.Errorf("register filter %q: %w", name, err)
		}
	}
	return nil
}

func filterFromJSON(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.IsError() {
		return in
	}
	if err := params.Take(); err != nil {
		return exec.AsValue(exec.ErrInvalidCall(err))
	}
	if !in.IsString() || in.String() == "" {
		return exec.AsValue(exec.ErrInvalidCall(fmt.Errorf("filter 'fromjson' was passed an empty or non-string type")))
	}
	// UseNumber keeps numbers as their literal representation instead of
	// rendering float64 forms like "8080.0"
	decoder := json.NewDecoder(strings.NewReader(in.String()))
	decoder.UseNumber()
	var object any
	if err := decoder.Decode(&object); err != nil {
		return exec.AsValue(fmt.Errorf("failed to unmarshal %s: %s", in.String(), err))
	}
	if _, err := decoder.Token(); err != io.EOF {
		return exec.AsValue(fmt.Errorf("failed to unmarshal %s: unexpected trailing data", in.String()))
	}
	return exec.AsValue(object)
}

func filterSplit(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.IsError() {
		return in
	}
	if !in.IsString() {
		return exec.AsValue(exec.ErrInvalidCall(fmt.Errorf("filter 'split' was passed a non-string type")))
	}
	var delimiter string
	if err := params.Take(
		exec.PositionalArgument("delimiter", nil, exec.StringArgument(&delimiter)),
	); err != nil {
		return exec.AsValue(exec.ErrInvalidCall(err))
	}
	list := strings.Split(in.String(), delimiter)
	out := make([]any, len(list))
	for index, item := range list {
		out[index] = item
	}
	return exec.AsValue(out)
}

func filterUnset(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.IsError() {
		return in
	}
	if !in.IsDict() {
		return exec.AsValue(exec.ErrInvalidCall(fmt.Errorf("filter 'unset' was passed a non-dict type")))
	}
	var toRemove string
	if err := params.Take(
		exec.PositionalArgument("key", nil, exec.StringArgument(&toRemove)),
	); err != nil {
		return exec.AsValue(exec.ErrInvalidCall(err))
	}
	out := map[string]any{}
	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		if key.String() != toRemove {
			out[key.String()] = value.Interface()
		}
		return true
	}, func() {})
	return exec.AsValue(out)
}
