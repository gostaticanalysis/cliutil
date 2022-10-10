package cliutil

import (
	"errors"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	// ErrNotFound indicates the object or the type could not be found
	ErrNotFound = errors.New("not found")
)

// DefaultConfig is the default value of [Config].
var DefaultConfig = &Config{
	Packages: &packages.Config{
		Mode: packages.NeedTypes,
	},
}

// TypeOf is wrapper of DefaultConfig.TypeOf(name).
func TypeOf(name string) (types.Type, error) {
	return DefaultConfig.TypeOf(name)
}

// ObjectOf is wrapper of DefaultConfig.ObjectOf(name).
func ObjectOf(name string) (types.Object, error) {
	return DefaultConfig.ObjectOf(name)
}

// Split splits name into three sections.
// The first section means a package name or a pre-declared identifier.
// The second section means an object of package.
// The third section means a field or a method.
// The fourth return value indicates whether name has "*" prefix or not.
func Split(name string) (first, second, third string, ptr bool) {

	slashed := strings.Split(name, "/")
	prefix := strings.Join(slashed[:len(slashed)-1], "/")
	splited := strings.Split(slashed[len(slashed)-1], ".")

	if prefix != "" {
		splited[0] = prefix + "/" + splited[0]
	}
	first = strings.TrimLeft(splited[0], "(")
	first = strings.TrimRight(first, ")")
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
