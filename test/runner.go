package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobuffalo/buffalo/meta"
	"github.com/gobuffalo/envy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var vendorPattern = "/vendor/"
var vendorRegex = regexp.MustCompile(vendorPattern)

type Runner struct {
	args []string
}

func NewRunner(args []string) Runner {
	args = removeFlag("--force-migrations", args)
	return Runner{
		args: args,
	}
}

func (tr Runner) Run() error {
	return testRunner(tr.args)
}

func testRunner(args []string) error {
	var mFlag bool
	var query string
	cargs := []string{}
	pargs := []string{}
	var larg string
	for i, a := range args {
		switch a {
		case "-run", "-m":
			query = args[i+1]
			mFlag = true
		case "-v":
			cargs = append(cargs, "-v")
		default:
			if larg != "-run" && larg != "-m" {
				pargs = append(pargs, a)
			}
		}
		larg = a
	}

	cmd := newTestCmd(cargs)
	if mFlag {
		return mFlagRunner{
			query: query,
			args:  cargs,
			pargs: pargs,
		}.Run()
	}

	pkgs, err := testPackages(pargs)
	if err != nil {
		return errors.WithStack(err)
	}
	cmd.Args = append(cmd.Args, pkgs...)
	logrus.Info(strings.Join(cmd.Args, " "))
	return cmd.Run()
}

type mFlagRunner struct {
	query string
	args  []string
	pargs []string
}

func (m mFlagRunner) Run() error {
	app := meta.New(".")
	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)

	pkgs, err := testPackages(m.pargs)
	if err != nil {
		return errors.WithStack(err)
	}
	var errs bool
	for _, p := range pkgs {
		os.Chdir(pwd)
		if p == app.PackagePkg {
			continue
		}
		p = strings.TrimPrefix(p, app.PackagePkg+string(filepath.Separator))
		os.Chdir(p)

		cmd := newTestCmd(m.args)
		if hasTestify(p) {
			cmd.Args = append(cmd.Args, "-testify.m", m.query)
		} else {
			cmd.Args = append(cmd.Args, "-run", m.query)
		}
		logrus.Info(strings.Join(cmd.Args, " "))
		if err := cmd.Run(); err != nil {
			errs = true
		}
	}
	if errs {
		return errors.New("errors running tests")
	}
	return nil
}

func hasTestify(p string) bool {
	cmd := exec.Command("go", "test", "-thisflagdoesntexist")
	b, _ := cmd.Output()
	return bytes.Contains(b, []byte("-testify.m"))
}

func testPackages(givenArgs []string) ([]string, error) {
	// If there are args, then assume these are the packages to test.
	//
	// Instead of always returning all packages from 'go list ./...', just
	// return the given packages in this case
	if len(givenArgs) > 0 {
		return givenArgs, nil
	}
	args := []string{}
	out, err := exec.Command(envy.Get("GO_BIN", "go"), "list", "./...").Output()
	if err != nil {
		return args, err
	}
	pkgs := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	for _, p := range pkgs {
		if !vendorRegex.Match(p) {
			args = append(args, string(p))
		}
	}
	return args, nil
}

func newTestCmd(args []string) *exec.Cmd {
	cargs := []string{"test", "-p", "1"}
	app := meta.New(".")
	cargs = append(cargs, "-tags", app.BuildTags("development").String())
	cargs = append(cargs, args...)
	cmd := exec.Command(envy.Get("GO_BIN", "go"), cargs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func removeFlag(flag string, args []string) []string {
	for i, v := range args {
		if v != flag {
			continue
		}

		args = append(args[:i], args[i+1:]...)
		break
	}

	return args
}
