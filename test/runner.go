package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/buffalo/meta"
	"github.com/gobuffalo/envy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Runner struct {
	args  []string
	mFlag bool
	query string

	pargs []string
	cargs []string
}

func NewRunner(args []string) Runner {
	args = removeFlag("--force-migrations", args)
	return Runner{
		args: args,
	}
}

func (tr Runner) Run() error {
	tr.parseFlags()

	if tr.mFlag {
		return tr.RunM()
	}

	return tr.RunRegular()

}
func (tr Runner) RunRegular() error {
	cmd := tr.buildTestCmd(tr.cargs)
	pkgs, err := tr.testPackages(tr.pargs)
	if err != nil {
		return errors.WithStack(err)
	}

	cmd.Args = append(cmd.Args, pkgs...)
	logrus.Info(strings.Join(cmd.Args, " "))

	return cmd.Run()
}

func (tr Runner) RunM() error {
	app := meta.New(".")
	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)

	pkgs, err := tr.testPackages(tr.pargs)
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

		cmd := tr.buildTestCmd(tr.cargs)
		if hasTestify(p) {
			cmd.Args = append(cmd.Args, "-testify.m", tr.query)
		} else {
			cmd.Args = append(cmd.Args, "-run", tr.query)
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

func (tr Runner) testPackages(givenArgs []string) ([]string, error) {
	// If there are args, then assume these are the packages to test.
	//
	// Instead of always returning all packages from 'go list ./...', just
	// return the given packages in this case
	if len(givenArgs) > 0 {
		return givenArgs, nil
	}

	return findTestPackages()
}

func (tr Runner) buildTestCmd(args []string) *exec.Cmd {
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

func (tr Runner) parseFlags() {
	var larg string
	for i, a := range tr.args {
		switch a {
		case "-run", "-m":
			tr.query = tr.args[i+1]
			tr.mFlag = true
		case "-v":
			tr.cargs = append(tr.cargs, "-v")
		default:
			if larg != "-run" && larg != "-m" {
				tr.pargs = append(tr.pargs, a)
			}
		}
		larg = a
	}
}
