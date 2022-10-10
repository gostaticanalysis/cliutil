package cliutil_test

import (
	"bytes"
	"errors"
	"go/types"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gostaticanalysis/cliutil"
	"golang.org/x/tools/go/types/objectpath"
)

func TestTypeOf(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		want    string
		wantErr error
	}{
		{"int", "int", nil},
		{"*int", "*int", nil},
		{"testing.T", "testing.T", nil},
		{"*testing.T", "*testing.T", nil},
		{"(golang.org/x/tools/go/types/objectpath).Path", "golang.org/x/tools/go/types/objectpath.Path", nil},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			conf := cliutil.NewConfigInDir(dir)

			// for third party packages
			pkg, _, _, _ := cliutil.Split(tt.name)
			if first, _, _ := strings.Cut(pkg, "/"); strings.Contains(first, ".") {
				execCmd(t, dir, "go mod init example.com/cliutil")
				execCmd(t, dir, "go get "+pkg)
			}

			typ, err := conf.TypeOf(tt.name)

			switch {
			case tt.wantErr != nil && errors.Is(err, tt.wantErr):
				t.Skip("expected error:", err)
			case err != nil:
				t.Fatal("unexpected error:", err)
			}

			var buf bytes.Buffer
			types.WriteType(&buf, typ, nil)

			if got := buf.String(); got != tt.want {
				t.Errorf("want %s but got %s", tt.want, got)
			}
		})
	}
}

func TestObjectOf(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		want    string
		wantErr error
	}{
		{"int", "int", nil},
		{"*int", "", cliutil.ErrNotFound},
		{"testing.T", "(testing).T", nil},
		{"*testing.T", "", cliutil.ErrNotFound},
		{"(*testing.T).Fatal", "(testing).common.M6", nil},
		{"(net/http.Handler).ServeHTTP", "(net/http).Handler.UM0", nil},
		{"(golang.org/x/tools/go/types/objectpath).For", "(golang.org/x/tools/go/types/objectpath).For", nil},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			conf := cliutil.NewConfigInDir(dir)

			// for third party packages
			pkg, _, _, _ := cliutil.Split(tt.name)
			if first, _, _ := strings.Cut(pkg, "/"); strings.Contains(first, ".") {
				execCmd(t, dir, "go mod init example.com/cliutil")
				execCmd(t, dir, "go get "+pkg)
			}

			obj, err := conf.ObjectOf(tt.name)

			switch {
			case tt.wantErr != nil && errors.Is(err, tt.wantErr):
				t.Skip("expected error:", err)
			case err != nil:
				t.Fatal("unexpected error:", err)
			}

			got := obj.Name() // for pre decl objects
			if pkg := obj.Pkg(); pkg != nil {
				pkgpath := "(" + pkg.Path() + ")."
				path, err := objectpath.For(obj)
				if err != nil {
					t.Fatal("unexpected error:", err)
				}
				got = pkgpath + string(path)
			}

			if got != tt.want {
				t.Errorf("want %s but got %s", tt.want, got)
			}
		})
	}
}

func TestCurrentPackage(t *testing.T) {
	t.Parallel()

	const (
		noErr   = false
		withErr = true
	)

	cases := []struct {
		name    string
		wantErr bool
	}{
		{"a", noErr},
		{"example.com/cliutil", noErr},
		{"---", withErr},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			conf := cliutil.NewConfigInDir(dir)

			execCmd(t, dir, "go mod init "+tt.name)
			gofile := filepath.Join(dir, "a.go")
			pkgname := path.Base(tt.name)
			src := []byte("package " + pkgname)
			if err := os.WriteFile(gofile, src, 0o644); err != nil {
				t.Fatal("unexpected error:", err)
			}

			pkg, err := conf.CurrentPackage()
			switch {
			case tt.wantErr && err == nil:
				t.Fatal("expected error did not occur")
			case !tt.wantErr && err != nil:
				t.Fatal("unexpected error:", err)
			case err != nil:
				t.Skip("expected error:", err)
			}

			t.Log(t.Name(), pkg)

			if got := pkg.Path(); got != tt.name {
				t.Errorf("want %s but got %s", tt.name, got)
			}
		})
	}
}

func execCmd(t *testing.T, dir, cmd string) string {
	t.Helper()
	if cmd == "" {
		return ""
	}

	args := strings.Split(cmd, " ")
	var buf bytes.Buffer
	_cmd := exec.Command(args[0], args[1:]...)
	_cmd.Stdout = &buf
	_cmd.Stderr = &buf
	_cmd.Dir = dir
	var eerr *exec.Error
	err := _cmd.Run()
	if errors.As(err, &eerr) {
		t.Fatal(eerr)
	}
	return buf.String()
}
