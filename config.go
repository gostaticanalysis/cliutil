package cliutil

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Config has configurations of static analysis.
type Config struct {
	Packages *packages.Config
}

// NewConfigInDir creates [Config] and set dir to Packages.Dir.
// Packages.Mode will be same with DefaultConfig.Packages.Mode.
func NewConfigInDir(dir string) *Config {
	return &Config{
		Packages: &packages.Config{
			Mode: DefaultConfig.Packages.Mode,
			Dir:  dir,
		},
	}
}

// TypeOf returns the value of types.Type which represented by the name.
// If any type could not be found, TypeOf returns [ErrNotFound] as the second return value.
func (conf *Config) TypeOf(name string) (types.Type, error) {
	first, second, _, ptr := Split(name)

	if second == "" {
		obj := types.Universe.Lookup(first)
		if obj == nil {
			return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
		}

		typ := obj.Type()
		if ptr {
			typ = types.NewPointer(typ)
		}

		return typ, nil
	}

	pkg, err := conf.load(first)
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

// ObjectOf returns the value of types.Object which represented by the name.
// If any object could not be found, ObjectOf returns [ErrNotFound] as the second return value.
func (conf *Config) ObjectOf(name string) (types.Object, error) {
	first, second, third, ptr := Split(name)
	if second == "" {
		obj := types.Universe.Lookup(name)
		if obj == nil {
			return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
		}
		return obj, nil
	}

	if third == "" && ptr {
		return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
	}

	pkg, err := conf.load(first)
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

func (conf *Config) load(name string) (*types.Package, error) {
	pkgs, err := packages.Load(conf.Packages, name)
	if err != nil {
		return nil, err
	}

	return pkgs[0].Types, nil
}
