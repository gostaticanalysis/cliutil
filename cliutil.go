package cliutil

import (
	"errors"
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	ErrNotFound = errors.New("not found")
)

var DefaultConfig = &packages.Config{Mode: packages.NeedTypes}

func TypeOf(name string) (types.Type, error) {
	first, second, _, ptr := Split(name)

	if second == "" {
		obj := types.Universe.Lookup(name)
		if obj == nil {
			return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
		}

		typ := obj.Type()
		if ptr {
			typ = types.NewPointer(typ)
		}

		return typ, nil
	}

	pkg, err := load(first)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", first, err)
	}

	obj := pkg.Scope().Lookup(second)
	if obj == nil {
		return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
	}

	typ := obj.Type()
	if ptr {
		typ = types.NewPointer(typ)
	}

	return typ, nil
}

func ObjectOf(name string) (types.Object, error) {
	first, second, third, ptr := Split(name)
	if second == "" {
		obj := types.Universe.Lookup(name)
		if obj == nil {
			return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
		}
		return obj, nil
	}

	pkg, err := load(first)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", first, err)
	}

	obj := pkg.Scope().Lookup(second)
	if obj == nil {
		return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
	}
	if third == "" {
		return obj, nil
	}

	fieldOrMethod, _, _ := types.LookupFieldOrMethod(obj.Type(), ptr, pkg, third)
	if fieldOrMethod == nil {
		return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
	}

	return fieldOrMethod, nil
}

func load(name string) (*types.Package, error) {
	pkgs, err := packages.Load(DefaultConfig, name)
	if err != nil {
		return nil, err
	}

	return pkgs[0].Types, nil
}

func Split(name string) (first, second, third string, ptr bool) {

	slashed := strings.Split(name, "/")
	prefix := strings.Join(slashed[:len(slashed)-1], "/")
	splited := strings.Split(slashed[len(slashed)-1], ".")

	if prefix != "" {
		splited[0] = prefix + "/" + splited[0]
	}
	first = strings.TrimLeft(splited[0], "(")
	if strings.HasPrefix(first, "*") {
		first = first[1:]
		ptr = true
	}

	// builtin?
	if len(splited) == 1 {
		return first, "", "", ptr
	}

	second = strings.TrimRight(splited[1], ")")

	if len(splited) >= 3 {
		third = splited[2]
	}

	return first, second, third, ptr
}
